package query_test

import (
	"context"
	"testing"

	"github.com/ernesto2108/anvil/internal/dashboard/imports"
	"github.com/ernesto2108/anvil/internal/dashboard/query"
	"github.com/ernesto2108/anvil/internal/dashboard/testutil"
	"github.com/ernesto2108/anvil/internal/dashboard/writer"
	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// seedRunForQuery inserts a minimal run so file_edges foreign key is satisfied.
func seedRunForQuery(t *testing.T, w *writer.EventWriter, runID string) {
	t.Helper()
	ev, err := instrumentation.NewEvent(runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
		TaskID: "task-" + runID,
	})
	if err != nil {
		t.Fatalf("NewEvent(RunStart): %v", err)
	}
	if err := w.WriteEvent(ev); err != nil {
		t.Fatalf("WriteEvent(RunStart): %v", err)
	}
}

func TestListFileEdgesByRun_ReturnsEdges(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRunForQuery(t, w, "run-lfe-1")

	edges := []imports.Edge{
		{SourcePath: "internal/a/a.go", TargetPath: "internal/b/b.go", Relation: "imports"},
		{SourcePath: "internal/a/a.go", TargetPath: "internal/c/c.go", Relation: "imports"},
	}
	if err := w.WriteFileEdges("run-lfe-1", edges); err != nil {
		t.Fatalf("WriteFileEdges: %v", err)
	}

	result, err := r.ListFileEdgesByRun(ctx, "run-lfe-1")
	if err != nil {
		t.Fatalf("ListFileEdgesByRun: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("edges count: got %d, want 2", len(result))
	}
	// Verify first edge fields.
	if result[0].SourcePath != "internal/a/a.go" {
		t.Errorf("SourcePath: got %q, want %q", result[0].SourcePath, "internal/a/a.go")
	}
	if result[0].Relation != "imports" {
		t.Errorf("Relation: got %q, want %q", result[0].Relation, "imports")
	}
}

func TestListFileEdgesByRun_EmptySliceWhenNone(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRunForQuery(t, w, "run-lfe-2")

	result, err := r.ListFileEdgesByRun(ctx, "run-lfe-2")
	if err != nil {
		t.Fatalf("ListFileEdgesByRun: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil empty slice, got nil")
	}
	if len(result) != 0 {
		t.Errorf("edges count: got %d, want 0", len(result))
	}
}

func TestListFileEdgesByRun_EmptySliceForUnknownRun(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	ctx := context.Background()

	result, err := r.ListFileEdgesByRun(ctx, "no-such-run")
	if err != nil {
		t.Fatalf("ListFileEdgesByRun(unknown): %v", err)
	}
	if result == nil {
		t.Error("expected non-nil empty slice, got nil")
	}
	if len(result) != 0 {
		t.Errorf("edges count: got %d, want 0", len(result))
	}
}

func TestGetRunTaskOrigin_ReturnsOrigin(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRunForQuery(t, w, "run-gto-1")

	origin := imports.TaskOrigin{
		Source: "linear",
		URL:    "https://linear.app/team/issue/ENG-42",
		ID:     "ENG-42",
	}
	if err := w.WriteTaskOrigin("run-gto-1", origin); err != nil {
		t.Fatalf("WriteTaskOrigin: %v", err)
	}

	source, url, trackerID, err := r.GetRunTaskOrigin(ctx, "run-gto-1")
	if err != nil {
		t.Fatalf("GetRunTaskOrigin: %v", err)
	}
	if source != "linear" {
		t.Errorf("source: got %q, want %q", source, "linear")
	}
	if url != "https://linear.app/team/issue/ENG-42" {
		t.Errorf("url: got %q", url)
	}
	if trackerID != "ENG-42" {
		t.Errorf("trackerID: got %q, want %q", trackerID, "ENG-42")
	}
}

func TestGetRunTaskOrigin_EmptyStringsWhenNoOrigin(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	w := writer.New(db, 0)
	ctx := context.Background()

	seedRunForQuery(t, w, "run-gto-2")

	// Run exists but task_source/task_url/task_tracker_id were never set.
	source, url, trackerID, err := r.GetRunTaskOrigin(ctx, "run-gto-2")
	if err != nil {
		t.Fatalf("GetRunTaskOrigin(no origin): %v", err)
	}
	if source != "" {
		t.Errorf("source: got %q, want empty", source)
	}
	if url != "" {
		t.Errorf("url: got %q, want empty", url)
	}
	if trackerID != "" {
		t.Errorf("trackerID: got %q, want empty", trackerID)
	}
}

func TestGetRunTaskOrigin_NoErrorWhenRunNotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	r := query.NewReader(db)
	ctx := context.Background()

	source, url, trackerID, err := r.GetRunTaskOrigin(ctx, "no-such-run")
	if err != nil {
		t.Fatalf("GetRunTaskOrigin(no-such-run): %v", err)
	}
	if source != "" || url != "" || trackerID != "" {
		t.Errorf("expected empty strings for missing run, got source=%q url=%q trackerID=%q", source, url, trackerID)
	}
}
