package cli

import (
	"testing"

	"github.com/ernesto2108/anvil/internal/dashboard/testutil"
	"github.com/ernesto2108/anvil/internal/dashboard/writer"
	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// TestEmptyRunIDs_FlagsOnlyEmptyTopLevelRuns verifies the cleanup query:
// only top-level runs with zero prompts/tool_uses/tasks (and no children)
// are flagged. Real runs and orchestrator parents are preserved.
func TestEmptyRunIDs_FlagsOnlyEmptyTopLevelRuns(t *testing.T) {
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 0)

	mustWrite := func(runID, eventType string, payload any) {
		t.Helper()
		if err := w.WriteEvent(mustEmitEvent(t, runID, eventType, payload)); err != nil {
			t.Fatalf("WriteEvent(%s/%s): %v", runID, eventType, err)
		}
	}

	// Empty run — should be flagged.
	mustWrite("run-empty", instrumentation.EventRunStart, instrumentation.RunStartPayload{
		SessionID: "sess-empty",
	})

	// Run with a user prompt — should NOT be flagged.
	mustWrite("run-with-prompt", instrumentation.EventRunStart, instrumentation.RunStartPayload{
		SessionID: "sess-prompt",
	})
	mustWrite("run-with-prompt", instrumentation.EventUserPrompt, instrumentation.UserPromptPayload{
		Prompt: "do the thing",
	})

	// Run with a tool use — should NOT be flagged. Need an agent first
	// because tool_uses references agents implicitly through agent_id.
	mustWrite("run-with-tool", instrumentation.EventRunStart, instrumentation.RunStartPayload{
		SessionID: "sess-tool",
	})
	mustWrite("run-with-tool", instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
		AgentID: "a1", AgentRole: "direct",
	})
	mustWrite("run-with-tool", instrumentation.EventToolUse, instrumentation.ToolUsePayload{
		AgentID: "a1", ToolName: "Bash", Source: "native",
	})

	ids, err := emptyRunIDs(db)
	if err != nil {
		t.Fatalf("emptyRunIDs: %v", err)
	}

	if len(ids) != 1 {
		t.Fatalf("expected 1 empty run, got %d: %v", len(ids), ids)
	}
	if ids[0] != "run-empty" {
		t.Errorf("flagged %q; want run-empty", ids[0])
	}
}

// TestDeleteRuns_CascadesToChildren verifies that DELETE FROM runs cascades
// through the foreign keys to agents/events/etc., relying on the migrations
// declaring ON DELETE CASCADE.
func TestDeleteRuns_CascadesToChildren(t *testing.T) {
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 0)

	if err := w.WriteEvent(mustEmitEvent(t, "run-cascade", instrumentation.EventRunStart, instrumentation.RunStartPayload{
		SessionID: "sess-cascade",
	})); err != nil {
		t.Fatalf("WriteEvent run.start: %v", err)
	}
	if err := w.WriteEvent(mustEmitEvent(t, "run-cascade", instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
		AgentID: "agent-c", AgentRole: "direct",
	})); err != nil {
		t.Fatalf("WriteEvent agent.start: %v", err)
	}

	deleted, err := deleteRuns(db, []string{"run-cascade"})
	if err != nil {
		t.Fatalf("deleteRuns: %v", err)
	}
	if deleted != 1 {
		t.Errorf("deleted %d; want 1", deleted)
	}

	var agents int
	if err := db.QueryRow(`SELECT COUNT(*) FROM agents WHERE run_id = ?`, "run-cascade").Scan(&agents); err != nil {
		t.Fatalf("query agents: %v", err)
	}
	if agents != 0 {
		t.Errorf("expected agents to cascade-delete; got %d remaining", agents)
	}
}
