package writer_test

import (
	"testing"

	"github.com/ernesto2108/anvil/internal/dashboard/testutil"
	"github.com/ernesto2108/anvil/internal/dashboard/writer"
	"github.com/ernesto2108/anvil/internal/instrumentation"
)

func TestResolveRunBySession(t *testing.T) {
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 0)

	// No run exists → empty string, no error
	runID, err := w.ResolveRunBySession("sess-xxx")
	if err != nil {
		t.Fatalf("ResolveRunBySession: %v", err)
	}
	if runID != "" {
		t.Errorf("expected empty runID, got %q", runID)
	}

	// Create a run with session_id
	w.WriteEvent(mustEvent(t, "run-sess-1", instrumentation.EventRunStart, instrumentation.RunStartPayload{
		SessionID: "sess-123",
	}))

	runID, err = w.ResolveRunBySession("sess-123")
	if err != nil {
		t.Fatalf("ResolveRunBySession: %v", err)
	}
	if runID != "run-sess-1" {
		t.Errorf("expected %q, got %q", "run-sess-1", runID)
	}
}

func TestUpdateTaskDesc(t *testing.T) {
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 0)

	w.WriteEvent(mustEvent(t, "run-td", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))

	// Should set task_desc when empty
	if err := w.UpdateTaskDesc("run-td", "first prompt"); err != nil {
		t.Fatalf("UpdateTaskDesc: %v", err)
	}

	var desc string
	db.QueryRow("SELECT task_desc FROM runs WHERE id = ?", "run-td").Scan(&desc)
	if desc != "first prompt" {
		t.Errorf("task_desc: got %q, want %q", desc, "first prompt")
	}

	// Should NOT overwrite when already set
	w.UpdateTaskDesc("run-td", "second prompt")
	db.QueryRow("SELECT task_desc FROM runs WHERE id = ?", "run-td").Scan(&desc)
	if desc != "first prompt" {
		t.Errorf("task_desc should not change: got %q", desc)
	}
}

func TestComputeRunTotals(t *testing.T) {
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 0)

	w.WriteEvent(mustEvent(t, "run-totals", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))
	w.WriteEvent(mustEvent(t, "run-totals", instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
		AgentID: "a1", AgentRole: "pm",
	}))
	w.WriteEvent(mustEvent(t, "run-totals", instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
		AgentID: "a2", AgentRole: "dev",
	}))
	w.WriteEvent(mustEvent(t, "run-totals", instrumentation.EventFileTouched, instrumentation.FileTouchedPayload{
		Path: "/a.go", Operation: "create",
	}))
	w.WriteEvent(mustEvent(t, "run-totals", instrumentation.EventFileTouched, instrumentation.FileTouchedPayload{
		Path: "/b.go", Operation: "create",
	}))
	// End the run to set ended_at
	w.WriteEvent(mustEvent(t, "run-totals", instrumentation.EventRunEnd, instrumentation.RunEndPayload{
		Status: "success",
	}))

	if err := w.ComputeRunTotals("run-totals"); err != nil {
		t.Fatalf("ComputeRunTotals: %v", err)
	}

	var filesCount, agentsCount int
	db.QueryRow("SELECT files_touched, agents_count FROM runs WHERE id = ?", "run-totals").
		Scan(&filesCount, &agentsCount)
	if filesCount != 2 {
		t.Errorf("files_touched: got %d, want 2", filesCount)
	}
	if agentsCount != 2 {
		t.Errorf("agents_count: got %d, want 2", agentsCount)
	}
}

func TestCleanupStaleRuns(t *testing.T) {
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 0)

	// Create a run that appears "stuck" — insert via SQL with all required fields
	// and a started_at 30 minutes in the past.
	_, err := db.Exec(`
		INSERT INTO runs (id, task_id, task_desc, status, complexity, provider, started_at, project, branch)
		VALUES (?, '', '', 'running', '', '', strftime('%Y-%m-%dT%H:%M:%fZ', 'now', '-30 minutes'), '', '')`,
		"stale-run")
	if err != nil {
		t.Fatalf("insert stale run: %v", err)
	}

	// Also create a recent run that should NOT be cleaned
	w.WriteEvent(mustEvent(t, "fresh-run", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))

	cleaned, err := w.CleanupStaleRuns(10)
	if err != nil {
		t.Fatalf("CleanupStaleRuns: %v", err)
	}
	if cleaned != 1 {
		t.Errorf("cleaned: got %d, want 1", cleaned)
	}

	var status string
	db.QueryRow("SELECT status FROM runs WHERE id = ?", "stale-run").Scan(&status)
	if status != "abandoned" {
		t.Errorf("stale status: got %q, want %q", status, "abandoned")
	}

	// Fresh run should still be running
	db.QueryRow("SELECT status FROM runs WHERE id = ?", "fresh-run").Scan(&status)
	if status != "running" {
		t.Errorf("fresh status: got %q, want %q", status, "running")
	}
}

func TestBackfillProjects(t *testing.T) {
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 0)

	// Create a run without a project
	w.WriteEvent(mustEvent(t, "run-bf", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))
	// Add a file with a path that contains the project name
	w.WriteEvent(mustEvent(t, "run-bf", instrumentation.EventFileTouched, instrumentation.FileTouchedPayload{
		Path: "/Users/dev/projects/myapp/src/main.go", Operation: "modify",
	}))

	filled, err := w.BackfillProjects()
	if err != nil {
		t.Fatalf("BackfillProjects: %v", err)
	}
	if filled != 1 {
		t.Errorf("filled: got %d, want 1", filled)
	}

	var project string
	db.QueryRow("SELECT project FROM runs WHERE id = ?", "run-bf").Scan(&project)
	if project != "myapp" {
		t.Errorf("project: got %q, want %q", project, "myapp")
	}
}

func TestAgentStartedAt(t *testing.T) {
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 0)

	// Not found
	_, ok := w.AgentStartedAt("x", "y")
	if ok {
		t.Error("expected ok=false for non-existent agent")
	}

	w.WriteEvent(mustEvent(t, "run-at", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))
	w.WriteEvent(mustEvent(t, "run-at", instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
		AgentID: "a1", AgentRole: "pm",
	}))

	startedAt, ok := w.AgentStartedAt("run-at", "a1")
	if !ok {
		t.Error("expected ok=true for existing agent")
	}
	if startedAt == "" {
		t.Error("expected non-empty startedAt")
	}
}

func TestExtractProjectFromPath(t *testing.T) {
	// This tests the extractProjectFromPath function indirectly via BackfillProjects.
	// We cover different path patterns by inserting files with various path structures.

	tests := []struct {
		name     string
		filePath string
		wantProj string
	}{
		{"projects dir", "/Users/dev/projects/myapp/src/main.go", "myapp"},
		{"repos dir", "/home/user/repos/backend/cmd/main.go", "backend"},
		{"src dir", "/opt/src/service/handler.go", "service"},
		{"workspace dir", "/home/user/workspace/tool/main.go", "tool"},
		{"fallback 4th segment", "/a/b/c/fallback/x.go", "fallback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testutil.OpenTestDB(t)
			w := writer.New(db, 0)

			w.WriteEvent(mustEvent(t, "run-ep", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))
			w.WriteEvent(mustEvent(t, "run-ep", instrumentation.EventFileTouched, instrumentation.FileTouchedPayload{
				Path: tt.filePath, Operation: "create",
			}))

			w.BackfillProjects()

			var project string
			db.QueryRow("SELECT project FROM runs WHERE id = ?", "run-ep").Scan(&project)
			if project != tt.wantProj {
				t.Errorf("project: got %q, want %q", project, tt.wantProj)
			}
		})
	}
}

func TestActiveAgentID(t *testing.T) {
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 0)

	w.WriteEvent(mustEvent(t, "run-aa", instrumentation.EventRunStart, instrumentation.RunStartPayload{}))

	// No active agent
	if id := w.ActiveAgentID("run-aa"); id != "" {
		t.Errorf("expected empty, got %q", id)
	}

	w.WriteEvent(mustEvent(t, "run-aa", instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
		AgentID: "a1", AgentRole: "pm",
	}))

	if id := w.ActiveAgentID("run-aa"); id != "a1" {
		t.Errorf("expected %q, got %q", "a1", id)
	}

	// After agent ends, no active agent
	w.WriteEvent(mustEvent(t, "run-aa", instrumentation.EventAgentEnd, instrumentation.AgentEndPayload{
		AgentID: "a1", Status: "success",
	}))

	if id := w.ActiveAgentID("run-aa"); id != "" {
		t.Errorf("expected empty after agent end, got %q", id)
	}
}
