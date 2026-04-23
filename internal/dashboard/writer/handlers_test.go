package writer_test

import (
	"testing"

	"github.com/ernesto2108/anvil/internal/dashboard/testutil"
	"github.com/ernesto2108/anvil/internal/dashboard/writer"
	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// TestUpdateToolUseDuration_Success verifies that UpdateToolUseDuration sets
// duration_ms on the matching tool_uses row (most recent with duration_ms IS NULL
// for the given run_id + tool_name + agent_id).
func TestUpdateToolUseDuration_Success(t *testing.T) {
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 0)

	// Create run + agent so FK constraints are satisfied.
	w.WriteEvent(mustEvent(t, "run-dur", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))
	w.WriteEvent(mustEvent(t, "run-dur", instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
		AgentID: "agent-1", AgentRole: "dev",
	}))

	// Insert a tool_use row via the existing PreToolUse path.
	w.WriteEvent(mustEvent(t, "run-dur", instrumentation.EventToolUse, instrumentation.ToolUsePayload{
		AgentID:   "agent-1",
		ToolName:  "mcp__postgres__query",
		MCPServer: "postgres",
		Source:    "mcp",
	}))

	// duration_ms should be NULL before the UPDATE.
	var durationBefore *int64
	err := db.QueryRow(
		"SELECT duration_ms FROM tool_uses WHERE run_id = ? AND tool_name = ? AND agent_id = ?",
		"run-dur", "mcp__postgres__query", "agent-1",
	).Scan(&durationBefore)
	if err != nil {
		t.Fatalf("query before update: %v", err)
	}
	if durationBefore != nil {
		t.Fatalf("expected duration_ms NULL before update, got %d", *durationBefore)
	}

	// Call the function under test.
	if err := w.UpdateToolUseDuration("run-dur", "mcp__postgres__query", "agent-1", 1500); err != nil {
		t.Fatalf("UpdateToolUseDuration: %v", err)
	}

	// Verify duration_ms is now set.
	var durationAfter *int64
	err = db.QueryRow(
		"SELECT duration_ms FROM tool_uses WHERE run_id = ? AND tool_name = ? AND agent_id = ?",
		"run-dur", "mcp__postgres__query", "agent-1",
	).Scan(&durationAfter)
	if err != nil {
		t.Fatalf("query after update: %v", err)
	}
	if durationAfter == nil {
		t.Fatal("expected duration_ms to be set, got NULL")
	}
	if *durationAfter != 1500 {
		t.Errorf("duration_ms: got %d, want 1500", *durationAfter)
	}
}

// TestUpdateToolUseDuration_NoMatch verifies that UpdateToolUseDuration does NOT
// return an error when no matching row is found (best-effort, silent ignore).
func TestUpdateToolUseDuration_NoMatch(t *testing.T) {
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 0)

	// No run, no tool_uses — the UPDATE finds no matching row.
	err := w.UpdateToolUseDuration("nonexistent-run", "mcp__postgres__query", "agent-x", 500)
	if err != nil {
		t.Errorf("UpdateToolUseDuration with no matching row should return nil, got: %v", err)
	}
}
