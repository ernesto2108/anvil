package query

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

// ListFileEdgesByRun returns all dependency edges for a run, ordered by id ASC.
func (r *Reader) ListFileEdgesByRun(ctx context.Context, runID string) ([]entity.FileEdge, error) {
	const q = `
		SELECT source_path, target_path, relation, COALESCE(agent_id, '')
		FROM file_edges
		WHERE run_id = ?
		ORDER BY id ASC`

	rows, err := r.db.QueryContext(ctx, q, runID)
	if err != nil {
		return nil, fmt.Errorf("dashboard/query: ListFileEdgesByRun: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var results []entity.FileEdge
	for rows.Next() {
		var e entity.FileEdge
		if err := rows.Scan(&e.SourcePath, &e.TargetPath, &e.Relation, &e.AgentID); err != nil {
			return nil, fmt.Errorf("dashboard/query: ListFileEdgesByRun scan: %w", err)
		}
		results = append(results, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard/query: ListFileEdgesByRun rows: %w", err)
	}
	if results == nil {
		results = []entity.FileEdge{}
	}
	return results, nil
}

// GetRunTaskOrigin returns the task_source, task_url, and task_tracker_id for a run.
// Returns empty strings if the run has no detected task origin or does not exist.
func (r *Reader) GetRunTaskOrigin(ctx context.Context, runID string) (source, url, trackerID string, err error) {
	const q = `SELECT task_source, task_url, task_tracker_id FROM runs WHERE id = ?`

	row := r.db.QueryRowContext(ctx, q, runID)
	var taskSource, taskURL, taskTrackerID sql.NullString
	if err := row.Scan(&taskSource, &taskURL, &taskTrackerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", "", nil
		}
		return "", "", "", fmt.Errorf("dashboard/query: GetRunTaskOrigin: %w", err)
	}
	return taskSource.String, taskURL.String, taskTrackerID.String, nil
}
