package memory

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
)

// ── mock embedder ─────────────────────────────────────────────────────────────

type mockEmbedder struct {
	vec []float32
	err error
}

func (m mockEmbedder) Embed(_ context.Context, _ string) ([]float32, error) {
	return m.vec, m.err
}

// ── helpers ───────────────────────────────────────────────────────────────────

// insertRun adds a run row to the test DB (newTestDB already inserts "r_test").
func insertRun(t *testing.T, db *sql.DB, id string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO runs (id, started_at) VALUES (?, '2026-04-16T00:00:00Z')`, id)
	if err != nil {
		t.Fatalf("insertRun %s: %v", id, err)
	}
}

// insertDigestEmb inserts a digest via UpsertDigest.
func insertDigestEmb(t *testing.T, db *sql.DB, id, runID, project string, emb []float32) {
	t.Helper()
	d := Digest{
		ID:        id,
		RunID:     runID,
		Project:   project,
		Summary:   "s_" + id,
		ModelUsed: "test-model",
		Embedding: emb,
	}
	if err := UpsertDigest(context.Background(), db, d); err != nil {
		t.Fatalf("insertDigestEmb %s: %v", id, err)
	}
}

// ── well-known 2-D vectors ────────────────────────────────────────────────────
// query = [1, 0]
// vec10  = [1, 0]  → cosine([1,0]) = 1.000
// vecHi  = [2, 1]  → cosine([1,0]) ≈ 0.894
// vecDia = [1, 1]  → cosine([1,0]) ≈ 0.707
// vec01  = [0, 1]  → cosine([1,0]) = 0.000

var (
	vecQ   = []float32{1, 0}
	vec10  = []float32{1, 0}
	vecHi  = []float32{2, 1}
	vecDia = []float32{1, 1}
	vec01  = []float32{0, 1}
)

// ── tests ─────────────────────────────────────────────────────────────────────

// TestSearchSimilar_ReturnsAboveThreshold verifies that only digests whose
// cosine score meets the threshold are included in results.
func TestSearchSimilar_ReturnsAboveThreshold(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	const proj = "p"

	// score ≈ 1.0 → passes threshold 0.5
	insertDigestEmb(t, db, "d1", "r_test", proj, vec10)
	// score ≈ 0.707 → passes threshold 0.5
	insertRun(t, db, "r2")
	insertDigestEmb(t, db, "d2", "r2", proj, vecDia)
	// score = 0.0 → below threshold 0.5
	insertRun(t, db, "r3")
	insertDigestEmb(t, db, "d3", "r3", proj, vec01)

	results, err := SearchSimilar(ctx, db, mockEmbedder{vec: vecQ}, "q", proj, 0, 0.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	for _, r := range results {
		if r.Score < 0.5 {
			t.Errorf("result %s has score %.4f < threshold 0.5", r.Digest.ID, r.Score)
		}
	}
}

// TestSearchSimilar_SortedByScoreDescending verifies that results are ordered
// from highest to lowest cosine score.
func TestSearchSimilar_SortedByScoreDescending(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	const proj = "p"

	// Insert in a "wrong" order so a naive unsorted return would fail.
	insertDigestEmb(t, db, "d1", "r_test", proj, vecDia) // ≈ 0.707
	insertRun(t, db, "r2")
	insertDigestEmb(t, db, "d2", "r2", proj, vec10) // 1.000
	insertRun(t, db, "r3")
	insertDigestEmb(t, db, "d3", "r3", proj, vecHi) // ≈ 0.894

	results, err := SearchSimilar(ctx, db, mockEmbedder{vec: vecQ}, "q", proj, 0, 0.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("need at least 2 results to verify ordering, got %d", len(results))
	}
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("out-of-order: score[%d]=%.4f > score[%d]=%.4f",
				i, results[i].Score, i-1, results[i-1].Score)
		}
	}
}

// TestSearchSimilar_LimitTruncates verifies that when limit > 0 and
// len(results) > limit, results are truncated to exactly limit entries.
func TestSearchSimilar_LimitTruncates(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	const proj = "p"

	insertDigestEmb(t, db, "d1", "r_test", proj, vec10)
	insertRun(t, db, "r2")
	insertDigestEmb(t, db, "d2", "r2", proj, vec10)
	insertRun(t, db, "r3")
	insertDigestEmb(t, db, "d3", "r3", proj, vec10)
	insertRun(t, db, "r4")
	insertDigestEmb(t, db, "d4", "r4", proj, vec10)
	insertRun(t, db, "r5")
	insertDigestEmb(t, db, "d5", "r5", proj, vec10)

	results, err := SearchSimilar(ctx, db, mockEmbedder{vec: vecQ}, "q", proj, 2, 0.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2 (limit=2)", len(results))
	}
}

// TestSearchSimilar_LimitZeroNoTruncation verifies that limit=0 disables
// truncation and all matching results are returned.
func TestSearchSimilar_LimitZeroNoTruncation(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	const proj = "p"

	insertDigestEmb(t, db, "d1", "r_test", proj, vec10)
	insertRun(t, db, "r2")
	insertDigestEmb(t, db, "d2", "r2", proj, vec10)
	insertRun(t, db, "r3")
	insertDigestEmb(t, db, "d3", "r3", proj, vec10)

	results, err := SearchSimilar(ctx, db, mockEmbedder{vec: vecQ}, "q", proj, 0, 0.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3 (limit=0 must not truncate)", len(results))
	}
}

// TestSearchSimilar_LimitLargerThanResults verifies that when limit exceeds the
// number of matching results, all results are returned without panic or error.
func TestSearchSimilar_LimitLargerThanResults(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	const proj = "p"

	insertDigestEmb(t, db, "d1", "r_test", proj, vec10)
	insertRun(t, db, "r2")
	insertDigestEmb(t, db, "d2", "r2", proj, vec10)

	results, err := SearchSimilar(ctx, db, mockEmbedder{vec: vecQ}, "q", proj, 100, 0.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2 (limit=100 > 2 results)", len(results))
	}
}

// TestSearchSimilar_SkipsNilEmbedding verifies that digests with a nil
// embedding are silently skipped and never passed to CosineSimilarity.
func TestSearchSimilar_SkipsNilEmbedding(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	const proj = "p"

	insertDigestEmb(t, db, "d_nil", "r_test", proj, nil) // must be skipped
	insertRun(t, db, "r2")
	insertDigestEmb(t, db, "d_ok", "r2", proj, vec10)

	results, err := SearchSimilar(ctx, db, mockEmbedder{vec: vecQ}, "q", proj, 0, 0.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1 (nil embedding must be skipped)", len(results))
	}
	if results[0].Digest.ID != "d_ok" {
		t.Errorf("expected digest id 'd_ok', got %q", results[0].Digest.ID)
	}
}

// TestSearchSimilar_EmbedError verifies that an Embedder failure is wrapped
// with the expected sentinel prefix "embed query".
func TestSearchSimilar_EmbedError(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	emb := mockEmbedder{err: errors.New("backend unavailable")}
	_, err := SearchSimilar(ctx, db, emb, "q", "p", 0, 0.5)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "embed query") {
		t.Errorf("error %q does not contain 'embed query'", err.Error())
	}
}

// TestSearchSimilar_DBError verifies that a database failure after a successful
// embed is wrapped with the expected sentinel prefix "load digests".
func TestSearchSimilar_DBError(t *testing.T) {
	// Open and immediately close so any query returns "sql: database is closed".
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	_ = db.Close()

	ctx := context.Background()
	emb := mockEmbedder{vec: vecQ}
	_, err = SearchSimilar(ctx, db, emb, "q", "p", 0, 0.5)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "load digests") {
		t.Errorf("error %q does not contain 'load digests'", err.Error())
	}
}

// TestSearchSimilar_NoDigests verifies that an empty project returns a zero-
// length slice (may be nil) with no error.
func TestSearchSimilar_NoDigests(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// project "empty" has no digests — r_test belongs to project "" by default
	results, err := SearchSimilar(ctx, db, mockEmbedder{vec: vecQ}, "q", "empty", 0, 0.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("len(results) = %d, want 0", len(results))
	}
}

// TestSearchSimilar_AllBelowThreshold verifies that when every digest scores
// below the threshold, results is empty (nil) with no error.
func TestSearchSimilar_AllBelowThreshold(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	const proj = "p"

	// vec01 ⊥ vecQ → cosine = 0.0, well below threshold 0.5
	insertDigestEmb(t, db, "d1", "r_test", proj, vec01)
	insertRun(t, db, "r2")
	insertDigestEmb(t, db, "d2", "r2", proj, vec01)

	results, err := SearchSimilar(ctx, db, mockEmbedder{vec: vecQ}, "q", proj, 0, 0.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("len(results) = %d, want 0 (all below threshold)", len(results))
	}
}
