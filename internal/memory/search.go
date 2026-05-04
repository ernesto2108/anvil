package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
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

// SearchKeyword runs a BM25 full-text search over fts_digests for the given
// query string, filtered by project. FTS5 MATCH syntax is supported: exact
// phrases ("sqlite-vec"), prefix (BLT*), boolean (AND/OR), field-scoped
// (summary: sqlite). Results are ordered by BM25 rank (most relevant first).
//
// The query is sanitized before being passed to MATCH: double-quotes are
// stripped to avoid breaking the FTS5 grammar when the caller passes raw
// identifiers like function names or task IDs.
func SearchKeyword(ctx context.Context, db *sql.DB, query, project string, limit int) ([]SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return nil, nil
	}
	// Strip characters that break FTS5 MATCH grammar: bare double-quotes
	// and unbalanced parens are the most common source of parse errors when
	// callers pass raw identifier strings.
	safe := strings.NewReplacer(`"`, ``, `(`, ``, `)`, ``).Replace(query)

	k := limit
	if k <= 0 {
		k = 1000
	}

	const q = `
		SELECT d.id, d.run_id, d.project, d.summary, d.decisions, d.edge_cases,
		       d.errors, d.embedding, d.model_used, d.source, d.created_at, d.updated_at,
		       fts.rank
		FROM fts_digests fts
		JOIN digests d ON d.id = fts.digest_id
		WHERE fts_digests MATCH ?
		  AND fts.project = ?
		ORDER BY fts.rank
		LIMIT ?`

	rows, err := db.QueryContext(ctx, q, safe, project, k)
	if err != nil {
		return nil, fmt.Errorf("memory: fts_digests match: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []SearchResult
	for rows.Next() {
		var (
			id        string
			runID     string
			proj      string
			summary   string
			decisions string
			edgeCases string
			errs      string
			embBlob   []byte
			modelUsed string
			source    string
			createdAt string
			updatedAt string
			rank      float64
		)
		if err := rows.Scan(&id, &runID, &proj, &summary, &decisions, &edgeCases, &errs,
			&embBlob, &modelUsed, &source, &createdAt, &updatedAt, &rank); err != nil {
			return nil, fmt.Errorf("memory: scan fts row: %w", err)
		}
		d := Digest{
			ID: id, RunID: runID, Project: proj, Summary: summary,
			ModelUsed: modelUsed, Source: source,
		}
		_ = json.Unmarshal([]byte(decisions), &d.Decisions)
		_ = json.Unmarshal([]byte(edgeCases), &d.EdgeCases)
		_ = json.Unmarshal([]byte(errs), &d.Errors)
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

		// FTS5 rank is negative BM25 (more negative = more relevant).
		// Normalize to [0,1] using a simple sigmoid so callers get a
		// score that is directionally consistent with SearchSimilar.
		score := 1.0 / (1.0 + (-rank))
		results = append(results, SearchResult{Digest: d, Score: score})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("memory: iterate fts rows: %w", err)
	}
	return results, nil
}

// SearchHybrid combines vector KNN (SearchSimilar) and BM25 keyword
// (SearchKeyword) results using Reciprocal Rank Fusion (RRF, k=60).
// Documents appearing in both rankings are boosted to the top.
// Requires a live Embedder for the vector leg; if embedding fails the
// function falls back to keyword-only results.
func SearchHybrid(ctx context.Context, db *sql.DB, emb Embedder, query, project string, limit int, threshold float64) ([]SearchResult, error) {
	const rrfK = 60.0

	vecResults, vecErr := SearchSimilar(ctx, db, emb, query, project, limit*2, threshold)
	kwResults, kwErr := SearchKeyword(ctx, db, query, project, limit*2)

	if vecErr != nil && kwErr != nil {
		return nil, fmt.Errorf("memory: hybrid search failed: vec=%w kw=%v", vecErr, kwErr)
	}

	type entry struct {
		digest   Digest
		rrfScore float64
	}
	byID := make(map[string]*entry)

	accumulateRRF := func(results []SearchResult) {
		for rank, r := range results {
			score := 1.0 / (rrfK + float64(rank+1))
			if e, ok := byID[r.Digest.ID]; ok {
				e.rrfScore += score
			} else {
				d := r.Digest
				byID[r.Digest.ID] = &entry{digest: d, rrfScore: score}
			}
		}
	}

	if vecErr == nil {
		accumulateRRF(vecResults)
	}
	if kwErr == nil {
		accumulateRRF(kwResults)
	}

	merged := make([]SearchResult, 0, len(byID))
	for _, e := range byID {
		merged = append(merged, SearchResult{Digest: e.digest, Score: e.rrfScore})
	}
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Score > merged[j].Score
	})

	if limit > 0 && len(merged) > limit {
		merged = merged[:limit]
	}
	return merged, nil
}
