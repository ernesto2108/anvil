package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// SearchSimilar embeds the query and runs a cosine KNN search over vec_digests
// (sqlite-vec virtual table), filtering by project via the partition key.
// Results above the similarity threshold are returned, ordered by score
// descending. The threshold uses similarity (1 - cosine distance) to keep the
// caller-facing semantics identical to the previous in-Go implementation.
func SearchSimilar(ctx context.Context, db *sql.DB, emb Embedder, query, project string, limit int, threshold float64) ([]SearchResult, error) {
	queryVec, err := emb.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("memory: embed query: %w", err)
	}
	queryBlob := EncodeEmbedding(queryVec)

	// vec0 requires a literal k. When the caller passes limit=0 ("no cap")
	// we fall back to a reasonably high k that covers all expected per-project
	// volumes; rows below the threshold are still discarded by the in-Go
	// filter below.
	k := limit
	if k <= 0 {
		k = 1000
	}

	const q = `
		SELECT d.id, d.run_id, d.project, d.summary, d.decisions, d.edge_cases,
		       d.errors, d.embedding, d.model_used, d.source, d.created_at, d.updated_at,
		       vd.distance
		FROM vec_digests vd
		JOIN digests d ON d.id = vd.digest_id
		WHERE vd.embedding MATCH ?
		  AND vd.project = ?
		  AND vd.k = ?
		ORDER BY vd.distance ASC`

	rows, err := db.QueryContext(ctx, q, queryBlob, project, k)
	if err != nil {
		return nil, fmt.Errorf("memory: vec_digests knn: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		d, distance, scanErr := scanDigestWithDistance(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		// cosine_similarity = 1 - cosine_distance. vec0 returns distance in
		// [0, 2]; clamp to keep the score in [-1, 1] when distances exceed 1
		// (orthogonal-or-worse vectors).
		score := 1.0 - distance
		if score >= threshold {
			results = append(results, SearchResult{Digest: d, Score: score})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("memory: iterate knn rows: %w", err)
	}

	return results, nil
}

// scanDigestWithDistance scans the joined row produced by SearchSimilar.
// Mirrors scanDigest but appends the vec_digests.distance column.
func scanDigestWithDistance(rows *sql.Rows) (Digest, float64, error) {
	var (
		id        string
		runID     string
		project   string
		summary   string
		decisions string
		edgeCases string
		errors    string
		embBlob   []byte
		modelUsed string
		source    string
		createdAt string
		updatedAt string
		distance  float64
	)
	if err := rows.Scan(&id, &runID, &project, &summary, &decisions, &edgeCases, &errors, &embBlob, &modelUsed, &source, &createdAt, &updatedAt, &distance); err != nil {
		return Digest{}, 0, fmt.Errorf("memory: scan digest+distance: %w", err)
	}

	d := Digest{
		ID:        id,
		RunID:     runID,
		Project:   project,
		Summary:   summary,
		ModelUsed: modelUsed,
		Source:    source,
	}
	_ = json.Unmarshal([]byte(decisions), &d.Decisions)
	_ = json.Unmarshal([]byte(edgeCases), &d.EdgeCases)
	_ = json.Unmarshal([]byte(errors), &d.Errors)
	if d.Decisions == nil {
		d.Decisions = []string{}
	}
	if d.EdgeCases == nil {
		d.EdgeCases = []string{}
	}
	if d.Errors == nil {
		d.Errors = []string{}
	}
	if embBlob != nil {
		if emb, decErr := DecodeEmbedding(embBlob); decErr == nil {
			d.Embedding = emb
		}
	}
	d.CreatedAt = parseFlexTime(createdAt)
	d.UpdatedAt = parseFlexTime(updatedAt)

	return d, distance, nil
}
