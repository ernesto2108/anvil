package query_test

import (
	"context"
	"testing"

	"github.com/ernesto2108/anvil/internal/dashboard/query"
	"github.com/ernesto2108/anvil/internal/dashboard/testutil"
	"github.com/ernesto2108/anvil/internal/dashboard/writer"
	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// seedRun creates a run with the given ID and optional events.
func seedRun(t *testing.T, w *writer.EventWriter, runID string, end bool) {
	t.Helper()
	ev, _ := instrumentation.NewEvent(runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
		TaskID:   "task-" + runID,
		Project:  "test-project",
		Provider: "anthropic",
	})
	w.WriteEvent(ev)

	if end {
		ev, _ = instrumentation.NewEvent(runID, instrumentation.EventRunEnd, instrumentation.RunEndPayload{
			Status: "success", DurationMs: 1000,
		})
		w.WriteEvent(ev)
	}
}

func TestListRuns(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRun(t, w, "run-a", true)
	seedRun(t, w, "run-b", true)
	seedRun(t, w, "run-c", false) // still running

	// List all
	runs, err := r.ListRuns(ctx, 10, 0, "", "", "", "")
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(runs) != 3 {
		t.Fatalf("ListRuns: got %d runs, want 3", len(runs))
	}

	// Filter by status
	runs, err = r.ListRuns(ctx, 10, 0, "running", "", "", "")
	if err != nil {
		t.Fatalf("ListRuns(running): %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("ListRuns(running): got %d, want 1", len(runs))
	}

	// Pagination
	runs, err = r.ListRuns(ctx, 1, 0, "", "", "", "")
	if err != nil {
		t.Fatalf("ListRuns(limit=1): %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("ListRuns(limit=1): got %d, want 1", len(runs))
	}

	// Filter by project
	runs, err = r.ListRuns(ctx, 10, 0, "", "", "", "test-project")
	if err != nil {
		t.Fatalf("ListRuns(project): %v", err)
	}
	if len(runs) != 3 {
		t.Errorf("ListRuns(project): got %d, want 3", len(runs))
	}
}

func TestGetRunSummary(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRun(t, w, "run-summary", true)

	summary, err := r.GetRunSummary(ctx, "run-summary")
	if err != nil {
		t.Fatalf("GetRunSummary: %v", err)
	}
	if summary == nil {
		t.Fatal("GetRunSummary: got nil")
	}
	if summary.ID != "run-summary" {
		t.Errorf("ID: got %q", summary.ID)
	}
	if summary.Status != "success" {
		t.Errorf("Status: got %q", summary.Status)
	}

	// Non-existent run
	summary, err = r.GetRunSummary(ctx, "no-exist")
	if err != nil {
		t.Fatalf("GetRunSummary(no-exist): %v", err)
	}
	if summary != nil {
		t.Error("expected nil for non-existent run")
	}
}

func TestListAgentsByRun(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRun(t, w, "run-agents", false)
	ev, _ := instrumentation.NewEvent("run-agents", instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
		AgentID: "a1", AgentRole: "pm", Sequence: 1,
	})
	w.WriteEvent(ev)
	ev, _ = instrumentation.NewEvent("run-agents", instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
		AgentID: "a2", AgentRole: "developer", Sequence: 2,
	})
	w.WriteEvent(ev)

	agents, err := r.ListAgentsByRun(ctx, "run-agents")
	if err != nil {
		t.Fatalf("ListAgentsByRun: %v", err)
	}
	if len(agents) != 2 {
		t.Fatalf("agents: got %d, want 2", len(agents))
	}
	// Should be ordered by sequence
	if agents[0].AgentRole != "pm" {
		t.Errorf("first agent: got %q, want pm", agents[0].AgentRole)
	}
	if agents[1].AgentRole != "developer" {
		t.Errorf("second agent: got %q, want developer", agents[1].AgentRole)
	}
}

func TestGetAgentDetail(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRun(t, w, "run-detail", false)
	ev, _ := instrumentation.NewEvent("run-detail", instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
		AgentID: "a1", AgentRole: "developer", Model: "claude-opus",
	})
	w.WriteEvent(ev)
	ev, _ = instrumentation.NewEvent("run-detail", instrumentation.EventAgentEnd, instrumentation.AgentEndPayload{
		AgentID: "a1", Status: "success", Output: "done",
		FilesTouched: []string{"/main.go"},
	})
	w.WriteEvent(ev)

	detail, files, err := r.GetAgentDetail(ctx, "run-detail", "a1")
	if err != nil {
		t.Fatalf("GetAgentDetail: %v", err)
	}
	if detail == nil {
		t.Fatal("detail is nil")
	}
	if detail.Output != "done" {
		t.Errorf("output: got %q", detail.Output)
	}
	if len(files) != 1 {
		t.Fatalf("files: got %d, want 1", len(files))
	}
	if files[0].Path != "/main.go" {
		t.Errorf("file path: got %q", files[0].Path)
	}

	// Non-existent agent
	detail, _, err = r.GetAgentDetail(ctx, "run-detail", "no-exist")
	if err != nil {
		t.Fatalf("GetAgentDetail(no-exist): %v", err)
	}
	if detail != nil {
		t.Error("expected nil for non-existent agent")
	}
}

func TestListFilesByRun(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRun(t, w, "run-files", false)
	ev, _ := instrumentation.NewEvent("run-files", instrumentation.EventFileTouched, instrumentation.FileTouchedPayload{
		Path: "/a.go", Operation: "create",
	})
	w.WriteEvent(ev)
	ev, _ = instrumentation.NewEvent("run-files", instrumentation.EventFileTouched, instrumentation.FileTouchedPayload{
		Path: "/b.go", Operation: "modify", Diff: "+new line",
	})
	w.WriteEvent(ev)

	files, err := r.ListFilesByRun(ctx, "run-files")
	if err != nil {
		t.Fatalf("ListFilesByRun: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("files: got %d, want 2", len(files))
	}
}

func TestListProjects(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	ev, _ := instrumentation.NewEvent("run-p1", instrumentation.EventRunStart, instrumentation.RunStartPayload{
		Project: "alpha",
	})
	w.WriteEvent(ev)
	ev, _ = instrumentation.NewEvent("run-p2", instrumentation.EventRunStart, instrumentation.RunStartPayload{
		Project: "beta",
	})
	w.WriteEvent(ev)
	ev, _ = instrumentation.NewEvent("run-p3", instrumentation.EventRunStart, instrumentation.RunStartPayload{
		Project: "alpha", // duplicate
	})
	w.WriteEvent(ev)

	projects, err := r.ListProjects(ctx)
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("projects: got %d, want 2 (alpha, beta)", len(projects))
	}
}

func TestErrorQueries(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRun(t, w, "run-err", false)
	ev, _ := instrumentation.NewEvent("run-err", instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
		AgentID: "a1", AgentRole: "dev",
	})
	w.WriteEvent(ev)
	ev, _ = instrumentation.NewEvent("run-err", instrumentation.EventAgentError, instrumentation.AgentErrorPayload{
		AgentID: "a1", Error: "null pointer dereference", DurationMs: 100,
	})
	w.WriteEvent(ev)

	// ListErrorGroups
	groups, err := r.ListErrorGroups(ctx, "", "")
	if err != nil {
		t.Fatalf("ListErrorGroups: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("error groups: got %d, want 1", len(groups))
	}
	if groups[0].OccurrenceCount != 1 {
		t.Errorf("occurrences: got %d, want 1", groups[0].OccurrenceCount)
	}

	// GetErrorGroup
	group, err := r.GetErrorGroup(ctx, groups[0].ID)
	if err != nil {
		t.Fatalf("GetErrorGroup: %v", err)
	}
	if group == nil {
		t.Fatal("GetErrorGroup returned nil")
	}

	// ListErrorGroupRuns
	runs, err := r.ListErrorGroupRuns(ctx, groups[0].ID)
	if err != nil {
		t.Fatalf("ListErrorGroupRuns: %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("error group runs: got %d, want 1", len(runs))
	}

	// GetErrorCounts
	counts, err := r.GetErrorCounts(ctx)
	if err != nil {
		t.Fatalf("GetErrorCounts: %v", err)
	}
	if counts.New != 1 {
		t.Errorf("new errors: got %d, want 1", counts.New)
	}

	// GetErrorTrend
	trend, err := r.GetErrorTrend(ctx, groups[0].ID)
	if err != nil {
		t.Fatalf("GetErrorTrend: %v", err)
	}
	if len(trend) != 14 {
		t.Errorf("trend points: got %d, want 14", len(trend))
	}

	// Filter by status
	groups, err = r.ListErrorGroups(ctx, "new", "")
	if err != nil {
		t.Fatalf("ListErrorGroups(new): %v", err)
	}
	if len(groups) != 1 {
		t.Errorf("new groups: got %d, want 1", len(groups))
	}

	// Search
	groups, err = r.ListErrorGroups(ctx, "", "null pointer")
	if err != nil {
		t.Fatalf("ListErrorGroups(search): %v", err)
	}
	if len(groups) != 1 {
		t.Errorf("search results: got %d, want 1", len(groups))
	}
}

func TestToolQueries(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRun(t, w, "run-tools", false)
	for i := 0; i < 3; i++ {
		ev, _ := instrumentation.NewEvent("run-tools", instrumentation.EventToolUse, instrumentation.ToolUsePayload{
			ToolName: "Bash", ToolInput: []byte(`{"command":"ls"}`),
		})
		w.WriteEvent(ev)
	}
	ev, _ := instrumentation.NewEvent("run-tools", instrumentation.EventToolUse, instrumentation.ToolUsePayload{
		ToolName: "Read", ToolInput: []byte(`{"file_path":"/a.go"}`),
	})
	w.WriteEvent(ev)

	usage, err := r.ListToolUsageByRun(ctx, "run-tools")
	if err != nil {
		t.Fatalf("ListToolUsageByRun: %v", err)
	}
	if len(usage) != 2 {
		t.Errorf("tool types: got %d, want 2", len(usage))
	}

	total, err := r.TotalToolUsesByRun(ctx, "run-tools")
	if err != nil {
		t.Fatalf("TotalToolUsesByRun: %v", err)
	}
	if total != 4 {
		t.Errorf("total tool uses: got %d, want 4", total)
	}

	details, err := r.ListToolUseDetailsByRun(ctx, "run-tools")
	if err != nil {
		t.Fatalf("ListToolUseDetailsByRun: %v", err)
	}
	if len(details) != 4 {
		t.Errorf("tool details: got %d, want 4", len(details))
	}
}

func TestDeleteRun(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRun(t, w, "run-del", true)

	if err := r.DeleteRun(ctx, "run-del"); err != nil {
		t.Fatalf("DeleteRun: %v", err)
	}

	summary, _ := r.GetRunSummary(ctx, "run-del")
	if summary != nil {
		t.Error("run should be deleted")
	}
}

func TestDeleteRuns(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRun(t, w, "run-d1", true)
	seedRun(t, w, "run-d2", true)
	seedRun(t, w, "run-d3", true)

	if err := r.DeleteRuns(ctx, []string{"run-d1", "run-d2"}); err != nil {
		t.Fatalf("DeleteRuns: %v", err)
	}

	runs, _ := r.ListRuns(ctx, 10, 0, "", "", "", "")
	if len(runs) != 1 {
		t.Errorf("remaining runs: got %d, want 1", len(runs))
	}
}

func TestTasksAndPrompts(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRun(t, w, "run-tp", false)

	ev, _ := instrumentation.NewEvent("run-tp", instrumentation.EventTaskCreated, instrumentation.TaskCreatedPayload{
		TaskID: "t1", Title: "task one",
	})
	w.WriteEvent(ev)
	ev, _ = instrumentation.NewEvent("run-tp", instrumentation.EventUserPrompt, instrumentation.UserPromptPayload{
		Prompt: "do something",
	})
	w.WriteEvent(ev)

	tasks, err := r.ListTasksByRun(ctx, "run-tp")
	if err != nil {
		t.Fatalf("ListTasksByRun: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("tasks: got %d, want 1", len(tasks))
	}

	prompts, err := r.ListPromptsByRun(ctx, "run-tp")
	if err != nil {
		t.Fatalf("ListPromptsByRun: %v", err)
	}
	if len(prompts) != 1 {
		t.Errorf("prompts: got %d, want 1", len(prompts))
	}
}

func TestMetricsQueries(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRun(t, w, "run-m1", false)

	// Add compaction and permission denied events
	ev, _ := instrumentation.NewEvent("run-m1", instrumentation.EventCompaction, instrumentation.CompactionPayload{Type: "auto"})
	w.WriteEvent(ev)
	ev, _ = instrumentation.NewEvent("run-m1", instrumentation.EventCompaction, instrumentation.CompactionPayload{Type: "auto"})
	w.WriteEvent(ev)

	compactions, err := r.CountCompactions(ctx, "run-m1")
	if err != nil {
		t.Fatalf("CountCompactions: %v", err)
	}
	if compactions != 2 {
		t.Errorf("compactions: got %d, want 2", compactions)
	}
}

func TestListActivityEvents(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRun(t, w, "run-act", false)
	ev, _ := instrumentation.NewEvent("run-act", instrumentation.EventFileTouched, instrumentation.FileTouchedPayload{
		AgentID: "a1", Path: "/x.go", Operation: "create",
	})
	w.WriteEvent(ev)

	events, err := r.ListActivityEvents(ctx, "run-act")
	if err != nil {
		t.Fatalf("ListActivityEvents: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("events: got %d, want 1", len(events))
	}
}

func TestListFilesByAgent(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRun(t, w, "run-fba", false)
	ev, _ := instrumentation.NewEvent("run-fba", instrumentation.EventFileTouched, instrumentation.FileTouchedPayload{
		AgentID: "a1", Path: "/a.go", Operation: "create",
	})
	w.WriteEvent(ev)
	ev, _ = instrumentation.NewEvent("run-fba", instrumentation.EventFileTouched, instrumentation.FileTouchedPayload{
		AgentID: "a2", Path: "/b.go", Operation: "modify",
	})
	w.WriteEvent(ev)

	files, err := r.ListFilesByAgent(ctx, "run-fba", "a1")
	if err != nil {
		t.Fatalf("ListFilesByAgent: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("files for a1: got %d, want 1", len(files))
	}
}

func TestCountPermissionDenied(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRun(t, w, "run-pd", false)
	ev, _ := instrumentation.NewEvent("run-pd", instrumentation.EventPermissionDenied, instrumentation.PermissionDeniedPayload{
		ToolName: "Bash",
	})
	w.WriteEvent(ev)

	count, err := r.CountPermissionDenied(ctx, "run-pd")
	if err != nil {
		t.Fatalf("CountPermissionDenied: %v", err)
	}
	if count != 1 {
		t.Errorf("permission denied: got %d, want 1", count)
	}
}

func TestGetTurnStats(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRun(t, w, "run-ts", false)
	// First prompt
	ev, _ := instrumentation.NewEvent("run-ts", instrumentation.EventUserPrompt, instrumentation.UserPromptPayload{
		Prompt: "fix the bug",
	})
	w.WriteEvent(ev)
	// Some activity
	ev, _ = instrumentation.NewEvent("run-ts", instrumentation.EventFileTouched, instrumentation.FileTouchedPayload{
		Path: "/a.go", Operation: "modify",
	})
	w.WriteEvent(ev)
	// Second prompt
	ev, _ = instrumentation.NewEvent("run-ts", instrumentation.EventUserPrompt, instrumentation.UserPromptPayload{
		Prompt: "add tests",
	})
	w.WriteEvent(ev)

	stats, err := r.GetTurnStats(ctx, "run-ts")
	if err != nil {
		t.Fatalf("GetTurnStats: %v", err)
	}
	if len(stats) != 2 {
		t.Errorf("turns: got %d, want 2", len(stats))
	}
}

func TestListChildRuns(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	// Parent run
	seedRun(t, w, "parent-run", false)

	// Child runs
	ev, _ := instrumentation.NewEvent("child-1", instrumentation.EventRunStart, instrumentation.RunStartPayload{
		ParentRunID: "parent-run",
	})
	w.WriteEvent(ev)
	ev, _ = instrumentation.NewEvent("child-2", instrumentation.EventRunStart, instrumentation.RunStartPayload{
		ParentRunID: "parent-run",
	})
	w.WriteEvent(ev)

	children, err := r.ListChildRuns(ctx, "parent-run")
	if err != nil {
		t.Fatalf("ListChildRuns: %v", err)
	}
	if len(children) != 2 {
		t.Errorf("children: got %d, want 2", len(children))
	}
}
