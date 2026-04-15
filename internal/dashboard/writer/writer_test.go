package writer_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/ernesto2108/anvil/internal/dashboard/testutil"
	"github.com/ernesto2108/anvil/internal/dashboard/writer"
	"github.com/ernesto2108/anvil/internal/instrumentation"
)

func newTestWriter(t *testing.T) (*writer.EventWriter, *sql.DB) {
	t.Helper()
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 500)
	return w, db
}

func mustEvent(t *testing.T, runID, eventType string, payload any) instrumentation.Event {
	t.Helper()
	ev, err := instrumentation.NewEvent(runID, eventType, payload)
	if err != nil {
		t.Fatalf("NewEvent(%s): %v", eventType, err)
	}
	return ev
}

func TestWriteEvent_RunStart(t *testing.T) {
	w, db := newTestWriter(t)

	ev := mustEvent(t, "run-001", instrumentation.EventRunStart, instrumentation.RunStartPayload{
		TaskID:          "task-1",
		TaskDescription: "test task",
		Complexity:      "simple",
		Provider:        "anthropic",
		SessionID:       "sess-abc",
		Project:         "anvil",
		Branch:          "main",
	})

	if err := w.WriteEvent(ev); err != nil {
		t.Fatalf("WriteEvent: %v", err)
	}

	var status, taskID, project, branch string
	err := db.QueryRow("SELECT status, task_id, project, branch FROM runs WHERE id = ?", "run-001").
		Scan(&status, &taskID, &project, &branch)
	if err != nil {
		t.Fatalf("query runs: %v", err)
	}
	if status != "running" {
		t.Errorf("status: got %q, want %q", status, "running")
	}
	if taskID != "task-1" {
		t.Errorf("task_id: got %q, want %q", taskID, "task-1")
	}
	if project != "anvil" {
		t.Errorf("project: got %q, want %q", project, "anvil")
	}
	if branch != "main" {
		t.Errorf("branch: got %q, want %q", branch, "main")
	}

	// Verify raw event was also inserted
	var evCount int
	db.QueryRow("SELECT COUNT(*) FROM events WHERE run_id = ?", "run-001").Scan(&evCount)
	if evCount != 1 {
		t.Errorf("events count: got %d, want 1", evCount)
	}
}

func TestWriteEvent_RunEnd(t *testing.T) {
	w, db := newTestWriter(t)

	w.WriteEvent(mustEvent(t, "run-002", instrumentation.EventRunStart, instrumentation.RunStartPayload{
		TaskID: "task-2",
	}))

	ev := mustEvent(t, "run-002", instrumentation.EventRunEnd, instrumentation.RunEndPayload{
		Status:          "success",
		DurationMs:      1500,
		TotalTokens:     5000,
		TotalFiles:      3,
		AgentsCompleted: 2,
		AgentsFailed:    0,
	})

	if err := w.WriteEvent(ev); err != nil {
		t.Fatalf("WriteEvent: %v", err)
	}

	var status string
	var durationMs int64
	var totalTokens, agentsCount int
	err := db.QueryRow("SELECT status, duration_ms, total_tokens, agents_count FROM runs WHERE id = ?", "run-002").
		Scan(&status, &durationMs, &totalTokens, &agentsCount)
	if err != nil {
		t.Fatalf("query runs: %v", err)
	}
	if status != "success" {
		t.Errorf("status: got %q, want %q", status, "success")
	}
	if durationMs != 1500 {
		t.Errorf("duration_ms: got %d, want 1500", durationMs)
	}
	if agentsCount != 2 {
		t.Errorf("agents_count: got %d, want 2", agentsCount)
	}
}

func TestWriteEvent_AgentLifecycle(t *testing.T) {
	w, db := newTestWriter(t)

	w.WriteEvent(mustEvent(t, "run-003", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))

	// Agent start
	w.WriteEvent(mustEvent(t, "run-003", instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
		AgentID:   "agent-pm",
		AgentRole: "pm",
		Sequence:  1,
		Model:     "claude-opus",
	}))

	var agentStatus, role string
	db.QueryRow("SELECT status, agent_role FROM agents WHERE run_id = ? AND agent_id = ?",
		"run-003", "agent-pm").Scan(&agentStatus, &role)
	if agentStatus != "running" {
		t.Errorf("agent status after start: got %q, want %q", agentStatus, "running")
	}
	if role != "pm" {
		t.Errorf("agent_role: got %q, want %q", role, "pm")
	}

	// Agent end with files
	w.WriteEvent(mustEvent(t, "run-003", instrumentation.EventAgentEnd, instrumentation.AgentEndPayload{
		AgentID:      "agent-pm",
		Status:       "success",
		DurationMs:   2000,
		Output:       "task completed",
		FilesTouched: []string{"/src/main.go"},
	}))

	db.QueryRow("SELECT status FROM agents WHERE run_id = ? AND agent_id = ?",
		"run-003", "agent-pm").Scan(&agentStatus)
	if agentStatus != "success" {
		t.Errorf("agent status after end: got %q, want %q", agentStatus, "success")
	}

	var fileCount int
	db.QueryRow("SELECT COUNT(*) FROM files WHERE run_id = ? AND agent_id = ?",
		"run-003", "agent-pm").Scan(&fileCount)
	if fileCount != 1 {
		t.Errorf("files from agent end: got %d, want 1", fileCount)
	}
}

func TestWriteEvent_AgentError_CreatesErrorGroup(t *testing.T) {
	w, db := newTestWriter(t)

	w.WriteEvent(mustEvent(t, "run-004", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))
	w.WriteEvent(mustEvent(t, "run-004", instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
		AgentID: "agent-dev", AgentRole: "developer",
	}))
	w.WriteEvent(mustEvent(t, "run-004", instrumentation.EventAgentError, instrumentation.AgentErrorPayload{
		AgentID: "agent-dev", Error: "compilation failed", DurationMs: 500,
	}))

	var status, errorMsg string
	db.QueryRow("SELECT status, error_msg FROM agents WHERE agent_id = ?", "agent-dev").
		Scan(&status, &errorMsg)
	if status != "failed" {
		t.Errorf("status: got %q, want %q", status, "failed")
	}
	if errorMsg != "compilation failed" {
		t.Errorf("error_msg: got %q", errorMsg)
	}

	var groupCount int
	db.QueryRow("SELECT COUNT(*) FROM error_groups").Scan(&groupCount)
	if groupCount != 1 {
		t.Errorf("error_groups: got %d, want 1", groupCount)
	}
}

func TestWriteEvent_FileTouched(t *testing.T) {
	w, db := newTestWriter(t)

	w.WriteEvent(mustEvent(t, "run-005", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))
	w.WriteEvent(mustEvent(t, "run-005", instrumentation.EventFileTouched, instrumentation.FileTouchedPayload{
		AgentID: "agent-x", Path: "/src/handler.go", Operation: "modify",
		Diff: "--- old\n+++ new",
	}))

	var path, operation string
	err := db.QueryRow("SELECT path, operation FROM files WHERE run_id = ?", "run-005").
		Scan(&path, &operation)
	if err != nil {
		t.Fatalf("query files: %v", err)
	}
	if path != "/src/handler.go" {
		t.Errorf("path: got %q", path)
	}
	if operation != "modify" {
		t.Errorf("operation: got %q", operation)
	}
}

func TestWriteEvent_QAScore(t *testing.T) {
	w, db := newTestWriter(t)

	w.WriteEvent(mustEvent(t, "run-006", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))
	w.WriteEvent(mustEvent(t, "run-006", instrumentation.EventQAScore, instrumentation.QAScorePayload{Score: 8.5}))

	var qaScore float64
	db.QueryRow("SELECT qa_score FROM runs WHERE id = ?", "run-006").Scan(&qaScore)
	if qaScore != 8.5 {
		t.Errorf("qa_score: got %f, want 8.5", qaScore)
	}
}

func TestWriteEvent_ToolUse(t *testing.T) {
	w, db := newTestWriter(t)

	w.WriteEvent(mustEvent(t, "run-007", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))
	w.WriteEvent(mustEvent(t, "run-007", instrumentation.EventToolUse, instrumentation.ToolUsePayload{
		AgentID: "agent-y", ToolName: "Bash", ToolInput: []byte(`{"command":"ls"}`),
	}))

	var toolName string
	db.QueryRow("SELECT tool_name FROM tool_uses WHERE run_id = ?", "run-007").Scan(&toolName)
	if toolName != "Bash" {
		t.Errorf("tool_name: got %q, want %q", toolName, "Bash")
	}
}

func TestWriteEvent_RunError_CreatesErrorGroup(t *testing.T) {
	w, db := newTestWriter(t)

	w.WriteEvent(mustEvent(t, "run-008", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))
	w.WriteEvent(mustEvent(t, "run-008", instrumentation.EventRunError, instrumentation.RunErrorPayload{
		ErrorReason: "context deadline exceeded",
	}))

	var status string
	db.QueryRow("SELECT status FROM runs WHERE id = ?", "run-008").Scan(&status)
	if status != "failed" {
		t.Errorf("status: got %q, want %q", status, "failed")
	}

	var groupCount int
	db.QueryRow("SELECT COUNT(*) FROM error_groups").Scan(&groupCount)
	if groupCount != 1 {
		t.Errorf("error_groups: got %d, want 1", groupCount)
	}
}

func TestWriteEvent_TaskLifecycle(t *testing.T) {
	w, db := newTestWriter(t)

	w.WriteEvent(mustEvent(t, "run-009", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))
	w.WriteEvent(mustEvent(t, "run-009", instrumentation.EventTaskCreated, instrumentation.TaskCreatedPayload{
		TaskID: "task-abc", Title: "implement feature",
	}))

	var taskStatus, title string
	db.QueryRow("SELECT status, title FROM tasks WHERE run_id = ? AND task_id = ?",
		"run-009", "task-abc").Scan(&taskStatus, &title)
	if taskStatus != "pending" {
		t.Errorf("task status: got %q, want %q", taskStatus, "pending")
	}

	w.WriteEvent(mustEvent(t, "run-009", instrumentation.EventTaskCompleted, instrumentation.TaskCompletedPayload{
		TaskID: "task-abc",
	}))

	db.QueryRow("SELECT status FROM tasks WHERE run_id = ? AND task_id = ?",
		"run-009", "task-abc").Scan(&taskStatus)
	if taskStatus != "completed" {
		t.Errorf("task status after complete: got %q, want %q", taskStatus, "completed")
	}
}

func TestWriteEvent_UserPrompt_SequenceNumbers(t *testing.T) {
	w, db := newTestWriter(t)

	w.WriteEvent(mustEvent(t, "run-010", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))
	w.WriteEvent(mustEvent(t, "run-010", instrumentation.EventUserPrompt, instrumentation.UserPromptPayload{Prompt: "first"}))
	w.WriteEvent(mustEvent(t, "run-010", instrumentation.EventUserPrompt, instrumentation.UserPromptPayload{Prompt: "second"}))

	var count int
	db.QueryRow("SELECT COUNT(*) FROM prompts WHERE run_id = ?", "run-010").Scan(&count)
	if count != 2 {
		t.Errorf("prompts count: got %d, want 2", count)
	}

	rows, _ := db.Query("SELECT sequence FROM prompts WHERE run_id = ? ORDER BY sequence", "run-010")
	defer rows.Close()
	var seqs []int
	for rows.Next() {
		var s int
		rows.Scan(&s)
		seqs = append(seqs, s)
	}
	if len(seqs) != 2 || seqs[0] != 1 || seqs[1] != 2 {
		t.Errorf("sequences: got %v, want [1,2]", seqs)
	}
}

func TestWriteEvent_Compaction_RawOnly(t *testing.T) {
	w, db := newTestWriter(t)

	w.WriteEvent(mustEvent(t, "run-011", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))
	w.WriteEvent(mustEvent(t, "run-011", instrumentation.EventCompaction, instrumentation.CompactionPayload{Type: "auto"}))

	var evCount int
	db.QueryRow("SELECT COUNT(*) FROM events WHERE run_id = ? AND event_type = ?",
		"run-011", instrumentation.EventCompaction).Scan(&evCount)
	if evCount != 1 {
		t.Errorf("compaction events: got %d, want 1", evCount)
	}
}

func TestWriteEvent_ErrorFingerprint_Regression(t *testing.T) {
	w, db := newTestWriter(t)

	// First occurrence
	w.WriteEvent(mustEvent(t, "run-r1", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))
	w.WriteEvent(mustEvent(t, "run-r1", instrumentation.EventRunError, instrumentation.RunErrorPayload{
		ErrorReason: "timeout waiting for response",
	}))

	// Get the group ID
	var groupID string
	db.QueryRow("SELECT id FROM error_groups LIMIT 1").Scan(&groupID)

	// Resolve the error
	w.UpdateErrorResolution(context.Background(), groupID, "resolved", "fixed", "abc123")

	// Same error reappears — regression
	w.WriteEvent(mustEvent(t, "run-r2", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))
	w.WriteEvent(mustEvent(t, "run-r2", instrumentation.EventRunError, instrumentation.RunErrorPayload{
		ErrorReason: "timeout waiting for response",
	}))

	var resStatus string
	var occurrences int
	db.QueryRow("SELECT resolution_status, occurrence_count FROM error_groups WHERE id = ?", groupID).
		Scan(&resStatus, &occurrences)
	if resStatus != "new" {
		t.Errorf("regression status: got %q, want %q", resStatus, "new")
	}
	if occurrences != 2 {
		t.Errorf("occurrences: got %d, want 2", occurrences)
	}
}

func TestDB_ReturnsUnderlyingDB(t *testing.T) {
	w, db := newTestWriter(t)
	if w.DB() != db {
		t.Error("DB() should return the underlying *sql.DB")
	}
}

func TestWriteEvent_PermissionDenied_RawOnly(t *testing.T) {
	w, db := newTestWriter(t)

	w.WriteEvent(mustEvent(t, "run-pd", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))
	w.WriteEvent(mustEvent(t, "run-pd", instrumentation.EventPermissionDenied, instrumentation.PermissionDeniedPayload{
		ToolName: "Bash",
	}))

	var evCount int
	db.QueryRow("SELECT COUNT(*) FROM events WHERE run_id = ? AND event_type = ?",
		"run-pd", instrumentation.EventPermissionDenied).Scan(&evCount)
	if evCount != 1 {
		t.Errorf("permission_denied events: got %d, want 1", evCount)
	}
}

func TestWriteEvent_RunEnd_UpdatesNotFound(t *testing.T) {
	w, _ := newTestWriter(t)

	// RunEnd without a RunStart should fail (no run to update)
	ev := mustEvent(t, "nonexistent-run", instrumentation.EventRunEnd, instrumentation.RunEndPayload{
		Status: "success",
	})
	err := w.WriteEvent(ev)
	if err == nil {
		t.Error("expected error when ending non-existent run")
	}
}

func TestRotate_PrunesOldRuns(t *testing.T) {
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 3) // Keep only 3 runs

	// Create 5 runs
	for i := 0; i < 5; i++ {
		runID := "run-rot-" + string(rune('a'+i))
		w.WriteEvent(mustEvent(t, runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{}))
		w.WriteEvent(mustEvent(t, runID, instrumentation.EventRunEnd, instrumentation.RunEndPayload{Status: "success"}))
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM runs").Scan(&count)
	if count > 3 {
		t.Errorf("runs after rotation: got %d, want <= 3", count)
	}
}
