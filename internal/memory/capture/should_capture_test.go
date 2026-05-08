package capture_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/ernesto2108/anvil/internal/memory/capture"
	"github.com/ernesto2108/anvil/internal/memory/transcript"
)

// openMinimalDB creates an in-memory SQLite DB with just the tables needed by
// ShouldCapture: tool_uses and files.
func openMinimalDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	const schema = `
		CREATE TABLE IF NOT EXISTS tool_uses (
			id      TEXT PRIMARY KEY,
			run_id  TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS files (
			id      TEXT PRIMARY KEY,
			run_id  TEXT NOT NULL
		);`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("create minimal schema: %v", err)
	}
	return db
}

func Test_ShouldCapture_ManualAlwaysTrue(t *testing.T) {
	// mode="manual" must return true even when DB is nil.
	got, err := capture.ShouldCapture(context.Background(), "some-run", transcript.TranscriptDigest{}, "manual", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("ShouldCapture manual = false, want true")
	}
}

func Test_ShouldCapture_DirectNoActivity(t *testing.T) {
	db := openMinimalDB(t)
	got, err := capture.ShouldCapture(context.Background(), "run-empty", transcript.TranscriptDigest{}, "direct", db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("ShouldCapture direct with 0 tool_uses + 0 files = true, want false")
	}
}

func Test_ShouldCapture_DirectWithToolUse(t *testing.T) {
	db := openMinimalDB(t)
	if _, err := db.Exec(`INSERT INTO tool_uses (id, run_id) VALUES ('tu1', 'run-active')`); err != nil {
		t.Fatalf("insert tool_use: %v", err)
	}

	got, err := capture.ShouldCapture(context.Background(), "run-active", transcript.TranscriptDigest{}, "direct", db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("ShouldCapture direct with 1 tool_use = false, want true")
	}
}

func Test_ShouldCapture_EmptyRunID(t *testing.T) {
	db := openMinimalDB(t)
	got, err := capture.ShouldCapture(context.Background(), "", transcript.TranscriptDigest{}, "direct", db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("ShouldCapture with empty runID = true, want false")
	}
}
