package writer

import (
	"fmt"

	"github.com/ernesto2108/anvil/internal/dashboard/imports"
)

// WriteTaskOrigin persists the task origin detection result for a run.
// It sets task_source, task_url, and task_tracker_id from the detected TaskOrigin.
// Only updates if the run's task_source is currently empty (idempotent on first write).
func (w *EventWriter) WriteTaskOrigin(runID string, origin imports.TaskOrigin) error {
	const q = `
		UPDATE runs
		SET task_source     = ?,
		    task_url        = ?,
		    task_tracker_id = ?
		WHERE id = ?`
	_, err := w.db.Exec(q, origin.Source, origin.URL, origin.ID, runID)
	if err != nil {
		return fmt.Errorf("dashboard/writer: WriteTaskOrigin: %w", err)
	}
	return nil
}

// WriteFileEdges persists dependency edges for a run.
// Uses INSERT OR IGNORE for idempotency (backed by unique index on run_id, source_path, target_path).
func (w *EventWriter) WriteFileEdges(runID string, edges []imports.Edge) error {
	const q = `INSERT OR IGNORE INTO file_edges (run_id, source_path, target_path, relation, agent_id) VALUES (?, ?, ?, ?, NULL)`
	for _, e := range edges {
		if _, err := w.db.Exec(q, runID, e.SourcePath, e.TargetPath, e.Relation); err != nil {
			return fmt.Errorf("dashboard/writer: WriteFileEdges: %w", err)
		}
	}
	return nil
}

// TouchedPathsForRun returns a set of relative paths touched in a run.
// The returned map has all paths as keys with value true.
func (w *EventWriter) TouchedPathsForRun(runID string) (map[string]bool, error) {
	rows, err := w.db.Query(`SELECT path FROM files WHERE run_id = ?`, runID)
	if err != nil {
		return nil, fmt.Errorf("dashboard/writer: TouchedPathsForRun: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	paths := make(map[string]bool)
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, fmt.Errorf("dashboard/writer: TouchedPathsForRun scan: %w", err)
		}
		paths[path] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard/writer: TouchedPathsForRun rows: %w", err)
	}
	return paths, nil
}
