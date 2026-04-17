package memory

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// newTestDB creates an in-memory SQLite database with the minimal schema
// required by digests (runs FK + digests table).
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open memory db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	const schema = `
CREATE TABLE runs (
    id         TEXT PRIMARY KEY,
    task_id    TEXT NOT NULL DEFAULT '',
    status     TEXT NOT NULL DEFAULT 'running',
    started_at TEXT NOT NULL,
    project    TEXT NOT NULL DEFAULT ''
);
CREATE TABLE digests (
    id          TEXT PRIMARY KEY,
    run_id      TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    project     TEXT NOT NULL,
    summary     TEXT NOT NULL,
    decisions   TEXT NOT NULL DEFAULT '[]',
    edge_cases  TEXT NOT NULL DEFAULT '[]',
    errors      TEXT NOT NULL DEFAULT '[]',
    embedding   BLOB,
    model_used  TEXT NOT NULL DEFAULT 'claude-haiku-4-5',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL,
    UNIQUE(run_id)
);`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO runs (id, started_at) VALUES ('r_test', '2026-04-16T00:00:00Z')`); err != nil {
		t.Fatalf("seed run: %v", err)
	}
	return db
}

func TestUpsertDigest_InsertThenUpdate(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	first := Digest{
		ID:        "d_1",
		RunID:     "r_test",
		Project:   "anvil",
		Summary:   "after agent 1",
		Decisions: []string{"chose option A"},
		EdgeCases: []string{"empty input"},
		Errors:    []string{},
		Embedding: []float32{0.1, 0.2, 0.3},
		ModelUsed: "mistral",
	}
	if err := UpsertDigest(ctx, db, first); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	got, err := GetDigestByRunID(ctx, db, "r_test")
	if err != nil || got == nil {
		t.Fatalf("read after insert: %v (got=%v)", err, got)
	}
	if got.Summary != "after agent 1" {
		t.Errorf("summary = %q, want 'after agent 1'", got.Summary)
	}

	// Update with richer content — same run_id must produce one row, not two.
	firstCreatedAt := got.CreatedAt
	time.Sleep(5 * time.Millisecond) // ensure updated_at diverges

	second := Digest{
		ID:        "d_2", // different id on purpose — existing row should keep d_1
		RunID:     "r_test",
		Project:   "anvil",
		Summary:   "after agent 2",
		Decisions: []string{"chose option A", "added retry"},
		EdgeCases: []string{"empty input", "timeout"},
		Errors:    []string{"migration dirty"},
		Embedding: []float32{0.4, 0.5, 0.6},
		ModelUsed: "claude-haiku-4-5",
	}
	if err := UpsertDigest(ctx, db, second); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	// Row count must be 1.
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM digests WHERE run_id = 'r_test'`).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("row count after upsert = %d, want 1", count)
	}

	got, err = GetDigestByRunID(ctx, db, "r_test")
	if err != nil || got == nil {
		t.Fatalf("read after update: %v (got=%v)", err, got)
	}

	// Content updated.
	if got.Summary != "after agent 2" {
		t.Errorf("summary = %q, want 'after agent 2'", got.Summary)
	}
	if len(got.Decisions) != 2 {
		t.Errorf("decisions len = %d, want 2", len(got.Decisions))
	}
	if len(got.Errors) != 1 || got.Errors[0] != "migration dirty" {
		t.Errorf("errors = %v, want ['migration dirty']", got.Errors)
	}
	if got.ModelUsed != "claude-haiku-4-5" {
		t.Errorf("model_used = %q, want 'claude-haiku-4-5'", got.ModelUsed)
	}

	// Identity preserved: id stays d_1, created_at stays the original.
	if got.ID != "d_1" {
		t.Errorf("id changed on update: got %q, want 'd_1'", got.ID)
	}
	if !got.CreatedAt.Equal(firstCreatedAt) {
		t.Errorf("created_at changed on update: got %v, want %v", got.CreatedAt, firstCreatedAt)
	}
	if !got.UpdatedAt.After(firstCreatedAt) {
		t.Errorf("updated_at not advanced: created=%v updated=%v", firstCreatedAt, got.UpdatedAt)
	}
}

func TestUpsertDigest_NilEmbedding(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	d := Digest{
		ID:        "d_x",
		RunID:     "r_test",
		Project:   "anvil",
		Summary:   "no embedding yet",
		ModelUsed: "mistral",
	}
	if err := UpsertDigest(ctx, db, d); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := GetDigestByRunID(ctx, db, "r_test")
	if err != nil || got == nil {
		t.Fatalf("read: %v (got=%v)", err, got)
	}
	if got.Embedding != nil {
		t.Errorf("embedding = %v, want nil", got.Embedding)
	}
}
