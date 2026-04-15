package query

import (
	"context"
	"fmt"
)

// CountCompactions returns the number of context.compacted events for a run.
func (r *Reader) CountCompactions(ctx context.Context, runID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM events WHERE run_id = ? AND event_type = 'context.compacted'`, runID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("query: contar compactaciones: %w", err)
	}
	return count, nil
}

// CountPermissionDenied returns the number of permission.denied events for a run.
func (r *Reader) CountPermissionDenied(ctx context.Context, runID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM events WHERE run_id = ? AND event_type = 'permission.denied'`, runID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("query: contar permission denied: %w", err)
	}
	return count, nil
}
