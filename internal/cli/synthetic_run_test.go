package cli

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/testutil"
)

func Test_EnsureSyntheticRun_CreatesRow(t *testing.T) {
	db := testutil.OpenTestDB(t)
	ts := time.Now()

	runID, err := EnsureSyntheticRun(context.Background(), db, "anvil", ts)
	if err != nil {
		t.Fatalf("EnsureSyntheticRun error: %v", err)
	}

	if !strings.HasPrefix(runID, "d_manual_") {
		t.Errorf("runID = %q, want prefix d_manual_", runID)
	}

	var taskSource string
	if err := db.QueryRow(`SELECT task_source FROM runs WHERE id = ?`, runID).Scan(&taskSource); err != nil {
		t.Fatalf("query inserted run: %v", err)
	}
	if taskSource != "manual" {
		t.Errorf("task_source = %q, want %q", taskSource, "manual")
	}
}

func Test_EnsureSyntheticRun_IdempotentOnConflict(t *testing.T) {
	db := testutil.OpenTestDB(t)
	ts := time.Now()

	// Insert once.
	runID1, err := EnsureSyntheticRun(context.Background(), db, "anvil", ts)
	if err != nil {
		t.Fatalf("first EnsureSyntheticRun error: %v", err)
	}

	// Manually insert the same ID to simulate a conflict.
	_, err = db.Exec(
		`INSERT INTO runs (id, task_id, task_desc, status, started_at, ended_at, project, task_source)
		 VALUES (?, ?, ?, 'success', ?, ?, ?, 'manual')
		 ON CONFLICT(id) DO NOTHING`,
		runID1, runID1, "duplicate", ts.UTC().Format(time.RFC3339Nano), ts.UTC().Format(time.RFC3339Nano), "anvil",
	)
	if err != nil {
		t.Fatalf("duplicate insert error: %v", err)
	}

	// A second call with the same timestamp should succeed (ON CONFLICT DO NOTHING).
	_, err = EnsureSyntheticRun(context.Background(), db, "anvil", ts)
	if err != nil {
		t.Errorf("second EnsureSyntheticRun error: %v (want idempotent)", err)
	}
}
