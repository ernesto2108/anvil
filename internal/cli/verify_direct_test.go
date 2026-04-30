package cli

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/testutil"
	"github.com/ernesto2108/anvil/internal/dashboard/writer"
	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// verifyHarness wires up an in-memory dashboard DB seeded with a run + agent
// so individual tests only declare the events that matter for the case.
type verifyHarness struct {
	db        *sql.DB
	w         *writer.EventWriter
	runID     string
	sessionID string
	agentID   string
	clock     time.Time
}

func newVerifyHarness(t *testing.T) *verifyHarness {
	t.Helper()
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 0)
	h := &verifyHarness{
		db:        db,
		w:         w,
		runID:     "run-verify-test",
		sessionID: "sess-verify-test",
		agentID:   "agent-direct",
		clock:     time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC),
	}
	h.writeAt(t, instrumentation.EventRunStart, instrumentation.RunStartPayload{
		SessionID: h.sessionID,
	})
	h.writeAt(t, instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
		AgentID:   h.agentID,
		AgentRole: "direct",
	})
	return h
}

// writeAt advances the harness clock by 1 second and writes an event at the
// new instant. Linear time prevents accidental ordering bugs in tests.
func (h *verifyHarness) writeAt(t *testing.T, eventType string, payload any) {
	t.Helper()
	h.clock = h.clock.Add(time.Second)
	ev, err := instrumentation.NewEvent(h.runID, eventType, payload)
	if err != nil {
		t.Fatalf("NewEvent(%s): %v", eventType, err)
	}
	ev.Timestamp = h.clock
	if err := h.w.WriteEvent(ev); err != nil {
		t.Fatalf("WriteEvent(%s): %v", eventType, err)
	}
}

func (h *verifyHarness) addPrompt(t *testing.T, prompt string) {
	h.writeAt(t, instrumentation.EventUserPrompt, instrumentation.UserPromptPayload{
		Prompt: prompt,
	})
}

func (h *verifyHarness) addEdit(t *testing.T, path string) {
	h.writeAt(t, instrumentation.EventFileTouched, instrumentation.FileTouchedPayload{
		AgentID:   h.agentID,
		Path:      path,
		Operation: "modify",
	})
}

func (h *verifyHarness) addBash(t *testing.T, command string, exitCode int) {
	cmdJSON, _ := json.Marshal(map[string]string{"command": command})
	h.writeAt(t, instrumentation.EventToolUse, instrumentation.ToolUsePayload{
		AgentID:   h.agentID,
		ToolName:  "Bash",
		ToolInput: cmdJSON,
		Source:    "native",
	})
	if err := h.w.UpdateToolUseResult(h.runID, "Bash", h.agentID, exitCode, ""); err != nil {
		t.Fatalf("UpdateToolUseResult: %v", err)
	}
}

func TestEvaluateVerify_NoSession_AllowsStop(t *testing.T) {
	h := newVerifyHarness(t)
	dec, err := evaluateVerify(h.db, "unknown-session")
	if err != nil {
		t.Fatalf("evaluateVerify: %v", err)
	}
	if dec != nil {
		t.Fatalf("expected nil decision for unknown session, got %+v", dec)
	}
}

func TestEvaluateVerify_NoPrompt_AllowsStop(t *testing.T) {
	h := newVerifyHarness(t)
	// run + agent exist but no prompt — can't tell turn boundaries.
	dec, err := evaluateVerify(h.db, h.sessionID)
	if err != nil {
		t.Fatalf("evaluateVerify: %v", err)
	}
	if dec != nil {
		t.Fatalf("expected nil decision when no prompt exists, got %+v", dec)
	}
}

func TestEvaluateVerify_NoEdits_AllowsStop(t *testing.T) {
	h := newVerifyHarness(t)
	h.addPrompt(t, "explica este endpoint")
	dec, err := evaluateVerify(h.db, h.sessionID)
	if err != nil {
		t.Fatalf("evaluateVerify: %v", err)
	}
	if dec != nil {
		t.Fatalf("expected nil decision with no edits, got %+v", dec)
	}
}

func TestEvaluateVerify_OnlyMarkdown_AllowsStop(t *testing.T) {
	h := newVerifyHarness(t)
	h.addPrompt(t, "actualiza el README")
	h.addEdit(t, "/repo/README.md")
	h.addEdit(t, "/repo/docs/guide.yaml")
	dec, err := evaluateVerify(h.db, h.sessionID)
	if err != nil {
		t.Fatalf("evaluateVerify: %v", err)
	}
	if dec != nil {
		t.Fatalf("expected nil decision for non-code edits, got %+v", dec)
	}
}

func TestEvaluateVerify_BypassPhraseSkipVerify(t *testing.T) {
	h := newVerifyHarness(t)
	h.addPrompt(t, "edita foo.go, skip verify por ahora")
	h.addEdit(t, "/repo/foo.go")
	dec, err := evaluateVerify(h.db, h.sessionID)
	if err != nil {
		t.Fatalf("evaluateVerify: %v", err)
	}
	if dec != nil {
		t.Fatalf("expected nil decision when prompt has bypass phrase, got %+v", dec)
	}
}

func TestEvaluateVerify_BypassPhraseSinVerify(t *testing.T) {
	h := newVerifyHarness(t)
	h.addPrompt(t, "Toca foo.go SIN verify, lo pruebo manual")
	h.addEdit(t, "/repo/foo.go")
	dec, err := evaluateVerify(h.db, h.sessionID)
	if err != nil {
		t.Fatalf("evaluateVerify: %v", err)
	}
	if dec != nil {
		t.Fatalf("expected nil decision for spanish bypass, got %+v", dec)
	}
}

func TestEvaluateVerify_GoEditNoLint_Blocks(t *testing.T) {
	h := newVerifyHarness(t)
	h.addPrompt(t, "agrega un campo a Run")
	h.addEdit(t, "/repo/internal/foo.go")
	dec, err := evaluateVerify(h.db, h.sessionID)
	if err != nil {
		t.Fatalf("evaluateVerify: %v", err)
	}
	if dec == nil {
		t.Fatalf("expected block decision; got nil")
	}
	if dec.Decision != "block" {
		t.Fatalf("expected decision=block, got %q", dec.Decision)
	}
	if !strings.Contains(dec.Reason, "Lint go") {
		t.Errorf("expected reason to mention Lint go; got %q", dec.Reason)
	}
	if !strings.Contains(dec.Reason, "Tests go") {
		t.Errorf("expected reason to mention Tests go; got %q", dec.Reason)
	}
}

func TestEvaluateVerify_GoEditLintPassedTestsMissing_Blocks(t *testing.T) {
	h := newVerifyHarness(t)
	h.addPrompt(t, "edita run.go")
	h.addEdit(t, "/repo/internal/run.go")
	h.addBash(t, "golangci-lint run ./...", 0)
	dec, err := evaluateVerify(h.db, h.sessionID)
	if err != nil {
		t.Fatalf("evaluateVerify: %v", err)
	}
	if dec == nil || dec.Decision != "block" {
		t.Fatalf("expected block; got %+v", dec)
	}
	if strings.Contains(dec.Reason, "Lint go") {
		t.Errorf("did not expect Lint go gap; reason: %q", dec.Reason)
	}
	if !strings.Contains(dec.Reason, "Tests go") {
		t.Errorf("expected Tests go gap; reason: %q", dec.Reason)
	}
}

func TestEvaluateVerify_GoEditBothPassed_AllowsStop(t *testing.T) {
	h := newVerifyHarness(t)
	h.addPrompt(t, "edita run.go")
	h.addEdit(t, "/repo/internal/run.go")
	h.addBash(t, "golangci-lint run ./...", 0)
	h.addBash(t, "go test ./internal/...", 0)
	dec, err := evaluateVerify(h.db, h.sessionID)
	if err != nil {
		t.Fatalf("evaluateVerify: %v", err)
	}
	if dec != nil {
		t.Fatalf("expected nil decision when lint+test both passed, got %+v", dec)
	}
}

func TestEvaluateVerify_LintRanBeforeEdit_Blocks(t *testing.T) {
	h := newVerifyHarness(t)
	h.addPrompt(t, "edita run.go")
	// lint runs FIRST, then the edit happens — lint is stale relative to it.
	h.addBash(t, "golangci-lint run ./...", 0)
	h.addBash(t, "go test ./...", 0)
	h.addEdit(t, "/repo/internal/run.go")
	dec, err := evaluateVerify(h.db, h.sessionID)
	if err != nil {
		t.Fatalf("evaluateVerify: %v", err)
	}
	if dec == nil || dec.Decision != "block" {
		t.Fatalf("expected block when verifications are stale; got %+v", dec)
	}
}

func TestEvaluateVerify_LintFailedExitNonZero_Blocks(t *testing.T) {
	h := newVerifyHarness(t)
	h.addPrompt(t, "edita run.go")
	h.addEdit(t, "/repo/internal/run.go")
	h.addBash(t, "golangci-lint run ./...", 1)
	h.addBash(t, "go test ./...", 0)
	dec, err := evaluateVerify(h.db, h.sessionID)
	if err != nil {
		t.Fatalf("evaluateVerify: %v", err)
	}
	if dec == nil || !strings.Contains(dec.Reason, "Lint go") {
		t.Fatalf("expected Lint go gap when lint failed; got %+v", dec)
	}
}

func TestEvaluateVerify_MultiStackPartiallyVerified_Blocks(t *testing.T) {
	h := newVerifyHarness(t)
	h.addPrompt(t, "ajustes en backend y front")
	h.addEdit(t, "/repo/internal/api.go")
	h.addEdit(t, "/repo/web/src/Foo.tsx")
	// Only the TS stack is verified.
	h.addBash(t, "pnpm exec eslint .", 0)
	h.addBash(t, "pnpm test", 0)
	dec, err := evaluateVerify(h.db, h.sessionID)
	if err != nil {
		t.Fatalf("evaluateVerify: %v", err)
	}
	if dec == nil || dec.Decision != "block" {
		t.Fatalf("expected block; got %+v", dec)
	}
	if !strings.Contains(dec.Reason, "Lint go") || !strings.Contains(dec.Reason, "Tests go") {
		t.Errorf("expected Go gaps; reason: %q", dec.Reason)
	}
	if strings.Contains(dec.Reason, "Lint ts") || strings.Contains(dec.Reason, "Tests ts") {
		t.Errorf("did not expect TS gaps; reason: %q", dec.Reason)
	}
}

func TestEvaluateVerify_PreviousTurnEditsIgnored(t *testing.T) {
	h := newVerifyHarness(t)
	// Turn 1: edit + verify both ran cleanly.
	h.addPrompt(t, "primer cambio")
	h.addEdit(t, "/repo/internal/a.go")
	h.addBash(t, "golangci-lint run", 0)
	h.addBash(t, "go test ./...", 0)
	// Turn 2: just exploration, no new edits.
	h.addPrompt(t, "explícame ese cambio")
	dec, err := evaluateVerify(h.db, h.sessionID)
	if err != nil {
		t.Fatalf("evaluateVerify: %v", err)
	}
	if dec != nil {
		t.Fatalf("expected nil decision when current turn has no edits, got %+v", dec)
	}
}

func TestEvaluateVerify_CombinedLintAndTestInOneBash(t *testing.T) {
	h := newVerifyHarness(t)
	h.addPrompt(t, "edita run.go")
	h.addEdit(t, "/repo/internal/run.go")
	// One Bash invocation that covers both lint and tests.
	h.addBash(t, "golangci-lint run ./... && go test ./...", 0)
	dec, err := evaluateVerify(h.db, h.sessionID)
	if err != nil {
		t.Fatalf("evaluateVerify: %v", err)
	}
	if dec != nil {
		t.Fatalf("expected nil decision when single Bash covers both; got %+v", dec)
	}
}
