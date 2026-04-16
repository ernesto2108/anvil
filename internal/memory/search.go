package memory

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
)

// SearchSimilar embeds the query, loads all project embeddings, computes cosine
// similarity, and returns the top results above the threshold.
func SearchSimilar(ctx context.Context, db *sql.DB, emb Embedder, query, project string, limit int, threshold float64) ([]SearchResult, error) {
	queryVec, err := emb.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("memory: embed query: %w", err)
	}

	digests, err := ListDigestsByProject(ctx, db, project)
	if err != nil {
		return nil, fmt.Errorf("memory: load digests: %w", err)
	}

	var results []SearchResult
	for _, d := range digests {
		if d.Embedding == nil {
			continue
		}
		score := CosineSimilarity(queryVec, d.Embedding)
		if score >= threshold {
			results = append(results, SearchResult{Digest: d, Score: score})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}
