package writer_test

import (
	"testing"

	"github.com/ernesto2108/anvil/internal/dashboard/imports"
	"github.com/ernesto2108/anvil/internal/dashboard/testutil"
	"github.com/ernesto2108/anvil/internal/dashboard/writer"
	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// seedRunForDeps inserts a minimal run row so that file inserts can reference it.
func seedRunForDeps(t *testing.T, w *writer.EventWriter, runID string) {
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

func TestWriteFileEdges_InsertsEdges(t *testing.T) {
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 0)

	seedRunForDeps(t, w, "run-fe-1")

	edges := []imports.Edge{
		{SourcePath: "internal/a/a.go", TargetPath: "internal/b/b.go", Relation: "imports"},
		{SourcePath: "internal/a/a.go", TargetPath: "internal/c/c.go", Relation: "imports"},
	}

	if err := w.WriteFileEdges("run-fe-1", edges); err != nil {
		t.Fatalf("WriteFileEdges: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM file_edges WHERE run_id = ?", "run-fe-1").Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	if count != 2 {
		t.Errorf("file_edges count: got %d, want 2", count)
	}
}

func TestWriteFileEdges_Idempotent(t *testing.T) {
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 0)

	seedRunForDeps(t, w, "run-fe-2")

	edges := []imports.Edge{
		{SourcePath: "internal/a/a.go", TargetPath: "internal/b/b.go", Relation: "imports"},
	}

	// Write the same edges twice — INSERT OR IGNORE guarantees one row.
	if err := w.WriteFileEdges("run-fe-2", edges); err != nil {
		t.Fatalf("WriteFileEdges first call: %v", err)
	}
	if err := w.WriteFileEdges("run-fe-2", edges); err != nil {
		t.Fatalf("WriteFileEdges second call: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM file_edges WHERE run_id = ?", "run-fe-2").Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	if count != 1 {
		t.Errorf("file_edges count after double-write: got %d, want 1 (idempotent)", count)
	}
}

func TestWriteFileEdges_EmptySlice(t *testing.T) {
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 0)

	seedRunForDeps(t, w, "run-fe-3")

	if err := w.WriteFileEdges("run-fe-3", nil); err != nil {
		t.Fatalf("WriteFileEdges(nil): %v", err)
	}
	if err := w.WriteFileEdges("run-fe-3", []imports.Edge{}); err != nil {
		t.Fatalf("WriteFileEdges([]): %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM file_edges WHERE run_id = ?", "run-fe-3").Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	if count != 0 {
		t.Errorf("file_edges count: got %d, want 0", count)
	}
}

func TestTouchedPathsForRun_ReturnsPaths(t *testing.T) {
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 0)

	seedRunForDeps(t, w, "run-tp-1")

	// Insert files via EventFileTouched.
	for _, path := range []string{"internal/a/a.go", "internal/b/b.go"} {
		ev, err := instrumentation.NewEvent("run-tp-1", instrumentation.EventFileTouched, instrumentation.FileTouchedPayload{
			Path: path, Operation: "modify",
		})
		if err != nil {
			t.Fatalf("NewEvent(FileTouched): %v", err)
		}
		if err := w.WriteEvent(ev); err != nil {
			t.Fatalf("WriteEvent(FileTouched): %v", err)
		}
	}

	paths, err := w.TouchedPathsForRun("run-tp-1")
	if err != nil {
		t.Fatalf("TouchedPathsForRun: %v", err)
	}
	if len(paths) != 2 {
		t.Errorf("paths count: got %d, want 2", len(paths))
	}
	if !paths["internal/a/a.go"] {
		t.Error("expected internal/a/a.go to be present")
	}
	if !paths["internal/b/b.go"] {
		t.Error("expected internal/b/b.go to be present")
	}
}

func TestTouchedPathsForRun_EmptyWhenNoFiles(t *testing.T) {
	db := testutil.OpenTestDB(t)
	w := writer.New(db, 0)

	seedRunForDeps(t, w, "run-tp-2")

	paths, err := w.TouchedPathsForRun("run-tp-2")
	if err != nil {
		t.Fatalf("TouchedPathsForRun: %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("paths count: got %d, want 0", len(paths))
	}
}
