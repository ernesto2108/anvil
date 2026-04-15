package query

import (
	"context"
	"fmt"

	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

// ListTasksByRun returns all tasks for a run, ordered by creation time.
func (r *Reader) ListTasksByRun(ctx context.Context, runID string) ([]entity.Task, error) {
	const q = `SELECT task_id, title, status, created_at, COALESCE(completed_at, '') FROM tasks WHERE run_id = ? ORDER BY id ASC`

	rows, err := r.db.QueryContext(ctx, q, runID)
	if err != nil {
		return nil, fmt.Errorf("query: consultar tasks del run %q: %w", runID, err)
	}
	defer rows.Close()

	var results []entity.Task
	for rows.Next() {
		var t entity.Task
		if err := rows.Scan(&t.TaskID, &t.Title, &t.Status, &t.CreatedAt, &t.CompletedAt); err != nil {
			return nil, fmt.Errorf("query: escanear task: %w", err)
		}
		results = append(results, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query: iterar tasks: %w", err)
	}
	return results, nil
}
