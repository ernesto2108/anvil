package cli

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"testing"
	"time"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"

	"github.com/ernesto2108/anvil/internal/instrumentation"
	"github.com/ernesto2108/anvil/internal/memory"
)

// checkpointVecOnce registers sqlite-vec exactly once for this test file so
// the in-memory DBs can host a vec_digests virtual table.
var checkpointVecOnce sync.Once

// --- test doubles ---

// fakeSummarizer records every Summarize call. When block is non-nil,
// Summarize waits for a receive on block before returning, so tests can
// control in-flight timing.
type fakeSummarizer struct {
	mu    sync.Mutex
	calls [][]string // each element = agent ids seen in one call

	block chan struct{} // nil = don't block
}

func (f *fakeSummarizer) Summarize(ctx context.Context, outs []memory.AgentOutput) (memory.DigestDraft, error) {
	if f.block != nil {
		select {
		case <-f.block:
		case <-ctx.Done():
			return memory.DigestDraft{}, ctx.Err()
		}
	}
	ids := make([]string, 0, len(outs))
	for _, o := range outs {
		ids = append(ids, o.AgentID)
	}
	f.mu.Lock()
	f.calls = append(f.calls, ids)
	f.mu.Unlock()
	return memory.DigestDraft{
		Summary:   "seen: " + strings.Join(ids, ","),
		Decisions: []string{},
		EdgeCases: []string{},
		Errors:    []string{},
	}, nil
}

func (f *fakeSummarizer) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

type fakeEmbedder struct{}

func (fakeEmbedder) Embed(_ context.Context, _ string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}

type nopSink struct{}

func (nopSink) Emit(_ instrumentation.Event) {}

// --- helpers ---

func newCheckpointTestDB(t *testing.T) *sql.DB {
	t.Helper()
	checkpointVecOnce.Do(sqlite_vec.Auto)
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
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
CREATE TABLE agents (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id     TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    agent_id   TEXT NOT NULL,
    agent_role TEXT NOT NULL,
    sequence   INTEGER,
    status     TEXT NOT NULL DEFAULT 'pending',
    output     TEXT
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
    model_used  TEXT NOT NULL DEFAULT 'fake',
    source      TEXT NOT NULL DEFAULT 'auto',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL,
    UNIQUE(run_id)
);
CREATE VIRTUAL TABLE vec_digests USING vec0(
    digest_id TEXT PRIMARY KEY,
    project TEXT partition key,
    embedding float[3] distance_metric=cosine
);`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("schema: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO runs (id, started_at) VALUES ('r_test', '2026-04-16T00:00:00Z')`); err != nil {
		t.Fatalf("seed run: %v", err)
	}
	return db
}

func seedAgent(t *testing.T, db *sql.DB, agentID, role, output string, sequence int) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO agents (run_id, agent_id, agent_role, sequence, status, output) VALUES ('r_test', ?, ?, ?, 'success', ?)`,
		agentID, role, sequence, output,
	)
	if err != nil {
		t.Fatalf("seed agent: %v", err)
	}
}

func agentEndEvent() instrumentation.Event {
	return instrumentation.Event{
		RunID:     "r_test",
		EventType: instrumentation.EventAgentEnd,
	}
}

func countDigests(t *testing.T, db *sql.DB) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM digests WHERE run_id = 'r_test'`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

// --- tests ---

// TestDigestCheckpointer_UpsertOnAgentEnd verifies that a single agent.end
// event triggers exactly one UPSERT into digests with the expected summary.
func TestDigestCheckpointer_UpsertOnAgentEnd(t *testing.T) {
	db := newCheckpointTestDB(t)
	seedAgent(t, db, "dev", "developer", "implemented foo", 1)

	sum := &fakeSummarizer{}
	cp := newDigestCheckpointer(context.Background(), nopSink{}, db, sum, fakeEmbedder{}, "fake-model", "r_test", "anvil")

	cp.Emit(agentEndEvent())
	cp.Wait()

	if got := sum.callCount(); got != 1 {
		t.Errorf("summarize calls = %d, want 1", got)
	}
	if n := countDigests(t, db); n != 1 {
		t.Errorf("digest rows = %d, want 1", n)
	}
	d, err := memory.GetDigestByRunID(context.Background(), db, "r_test")
	if err != nil || d == nil {
		t.Fatalf("get digest: err=%v d=%v", err, d)
	}
	if d.Summary != "seen: dev" {
		t.Errorf("summary = %q, want 'seen: dev'", d.Summary)
	}
	if d.Embedding == nil {
		t.Error("embedding should be set (fakeEmbedder returns 3 floats)")
	}
}

// TestDigestCheckpointer_UpdatesAcrossAgents verifies that a second agent.end
// (with more agent rows in the DB) overwrites the digest in place — one row,
// updated summary, id preserved, updated_at advanced.
func TestDigestCheckpointer_UpdatesAcrossAgents(t *testing.T) {
	db := newCheckpointTestDB(t)
	seedAgent(t, db, "dev", "developer", "implemented foo", 1)

	sum := &fakeSummarizer{}
	cp := newDigestCheckpointer(context.Background(), nopSink{}, db, sum, fakeEmbedder{}, "fake-model", "r_test", "anvil")

	cp.Emit(agentEndEvent())
	cp.Wait()

	d1, _ := memory.GetDigestByRunID(context.Background(), db, "r_test")
	if d1 == nil {
		t.Fatal("first digest missing")
	}
	firstID := d1.ID
	firstUpdated := d1.UpdatedAt

	// Simulate a second agent completing.
	seedAgent(t, db, "test", "tester", "added tests", 2)
	time.Sleep(5 * time.Millisecond) // ensure updated_at diverges

	cp.Emit(agentEndEvent())
	cp.Wait()

	if n := countDigests(t, db); n != 1 {
		t.Fatalf("digest rows after second agent = %d, want 1", n)
	}
	d2, _ := memory.GetDigestByRunID(context.Background(), db, "r_test")
	if d2 == nil {
		t.Fatal("second digest missing")
	}
	if d2.ID != firstID {
		t.Errorf("id changed: %q -> %q", firstID, d2.ID)
	}
	if d2.Summary != "seen: dev,test" {
		t.Errorf("summary = %q, want 'seen: dev,test'", d2.Summary)
	}
	if !d2.UpdatedAt.After(firstUpdated) {
		t.Errorf("updated_at did not advance: first=%v second=%v", firstUpdated, d2.UpdatedAt)
	}
}

// TestDigestCheckpointer_CoalescesRapidEvents verifies that when many
// agent.end events arrive while a checkpoint is in flight, the worker runs
// at most once more (not once per event). This protects against summarizer
// call storms during fast pipelines.
func TestDigestCheckpointer_CoalescesRapidEvents(t *testing.T) {
	db := newCheckpointTestDB(t)
	seedAgent(t, db, "dev", "developer", "implemented foo", 1)

	// Blocking summarizer: first call hangs until we release it.
	block := make(chan struct{})
	sum := &fakeSummarizer{block: block}
	cp := newDigestCheckpointer(context.Background(), nopSink{}, db, sum, fakeEmbedder{}, "fake-model", "r_test", "anvil")

	// First emit starts the worker; worker is now blocked in Summarize.
	cp.Emit(agentEndEvent())

	// Wait until the worker is actually in Summarize (running=true).
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		cp.mu.Lock()
		running := cp.running
		cp.mu.Unlock()
		if running {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Three more emits while the first is still running. Should all coalesce
	// into a single pending flag.
	cp.Emit(agentEndEvent())
	cp.Emit(agentEndEvent())
	cp.Emit(agentEndEvent())

	// Release the first call. The worker will loop once more (pending=true).
	// Unblock enough for both calls (close = unblocks all receives).
	close(block)

	cp.Wait()

	if got := sum.callCount(); got != 2 {
		t.Errorf("summarize calls = %d, want 2 (1 initial + 1 coalesced pending)", got)
	}
}

// TestDigestCheckpointer_ParentCtxCancels verifies that when the parent
// context is cancelled (simulating Ctrl+C), an in-flight checkpoint aborts
// quickly instead of blocking for the per-checkpoint timeout.
func TestDigestCheckpointer_ParentCtxCancels(t *testing.T) {
	db := newCheckpointTestDB(t)
	seedAgent(t, db, "dev", "developer", "implemented foo", 1)

	// Blocking summarizer that never gets released — only ctx cancellation
	// should unblock it.
	block := make(chan struct{}) // never closed
	sum := &fakeSummarizer{block: block}

	parentCtx, cancel := context.WithCancel(context.Background())
	cp := newDigestCheckpointer(parentCtx, nopSink{}, db, sum, fakeEmbedder{}, "fake-model", "r_test", "anvil")

	cp.Emit(agentEndEvent())

	// Give the worker a moment to enter Summarize.
	time.Sleep(20 * time.Millisecond)

	// Cancel parent. Wait() should return quickly.
	cancel()

	done := make(chan struct{})
	go func() {
		cp.Wait()
		close(done)
	}()

	select {
	case <-done:
		// good
	case <-time.After(2 * time.Second):
		t.Fatal("Wait() did not return within 2s after parent ctx cancellation")
	}

	// The shutdown guarantee is: Wait() returns promptly after parent ctx
	// cancellation. Whether a row was partially written is implementation
	// detail — what matters is no hang.
}
