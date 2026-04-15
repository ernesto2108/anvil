package query

import (
	"context"
	"fmt"

	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

// ListToolUsageByRun returns tool usage counts for a run, ordered by count DESC.
func (r *Reader) ListToolUsageByRun(ctx context.Context, runID string) ([]entity.ToolUsage, error) {
	const q = `SELECT tool_name, COUNT(*) AS cnt FROM tool_uses WHERE run_id = ? GROUP BY tool_name ORDER BY cnt DESC`

	rows, err := r.db.QueryContext(ctx, q, runID)
	if err != nil {
		return nil, fmt.Errorf("query: consultar tool_uses del run %q: %w", runID, err)
	}
	defer rows.Close()

	var results []entity.ToolUsage
	for rows.Next() {
		var t entity.ToolUsage
		if err := rows.Scan(&t.ToolName, &t.Count); err != nil {
			return nil, fmt.Errorf("query: escanear tool_uses: %w", err)
		}
		results = append(results, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query: iterar tool_uses: %w", err)
	}
	return results, nil
}

// TotalToolUsesByRun returns the total count of tool invocations for a run.
func (r *Reader) TotalToolUsesByRun(ctx context.Context, runID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tool_uses WHERE run_id = ?`, runID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("query: contar tool_uses: %w", err)
	}
	return count, nil
}

// ListToolUseDetailsByRun returns individual tool invocations for a run,
// filtered to tools that have a non-empty input.
func (r *Reader) ListToolUseDetailsByRun(ctx context.Context, runID string) ([]entity.ToolUseDetail, error) {
	const q = `SELECT tool_name, tool_input, agent_id, timestamp
	           FROM tool_uses
	           WHERE run_id = ? AND tool_input != ''
	           ORDER BY timestamp ASC`

	rows, err := r.db.QueryContext(ctx, q, runID)
	if err != nil {
		return nil, fmt.Errorf("query: consultar tool_use details del run %q: %w", runID, err)
	}
	defer rows.Close()

	var results []entity.ToolUseDetail
	for rows.Next() {
		var d entity.ToolUseDetail
		if err := rows.Scan(&d.ToolName, &d.ToolInput, &d.AgentID, &d.Timestamp); err != nil {
			return nil, fmt.Errorf("query: escanear tool_use detail: %w", err)
		}
		results = append(results, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query: iterar tool_use details: %w", err)
	}
	return results, nil
}
