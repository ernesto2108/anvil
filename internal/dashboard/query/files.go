package query

import (
	"context"
	"fmt"

	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

// ListFilesByRun returns all files touched in a run, ordered by id ASC.
func (r *Reader) ListFilesByRun(ctx context.Context, runID string) ([]entity.FileRecord, error) {
	const q = `
		SELECT path, operation, COALESCE(diff, ''), agent_id
		FROM files
		WHERE run_id = ?
		ORDER BY id ASC`

	rows, err := r.db.QueryContext(ctx, q, runID)
	if err != nil {
		return nil, fmt.Errorf("query: consultar files del run %q: %w", runID, err)
	}
	defer rows.Close()

	var results []entity.FileRecord
	for rows.Next() {
		var f entity.FileRecord
		if err := rows.Scan(&f.Path, &f.Operation, &f.Diff, &f.AgentID); err != nil {
			return nil, fmt.Errorf("query: escanear fila de files: %w", err)
		}
		results = append(results, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query: iterar filas de files: %w", err)
	}
	return results, nil
}

// ListFilesByAgent returns files for a specific agent within a run.
func (r *Reader) ListFilesByAgent(ctx context.Context, runID, agentID string) ([]entity.FileRecord, error) {
	const q = `
		SELECT path, operation, COALESCE(diff, ''), agent_id
		FROM files
		WHERE run_id = ? AND agent_id = ?
		ORDER BY id ASC`

	rows, err := r.db.QueryContext(ctx, q, runID, agentID)
	if err != nil {
		return nil, fmt.Errorf("query: consultar files del agente %q: %w", agentID, err)
	}
	defer rows.Close()

	var results []entity.FileRecord
	for rows.Next() {
		var f entity.FileRecord
		if err := rows.Scan(&f.Path, &f.Operation, &f.Diff, &f.AgentID); err != nil {
			return nil, fmt.Errorf("query: escanear fila de files: %w", err)
		}
		results = append(results, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query: iterar filas de files: %w", err)
	}
	return results, nil
}

// ListActivityEvents returns file.touched events with timestamps and agent attribution.
func (r *Reader) ListActivityEvents(ctx context.Context, runID string) ([]entity.ActivityEvent, error) {
	const q = `
		SELECT timestamp, COALESCE(json_extract(payload, '$.agent_id'), '')
		FROM events
		WHERE run_id = ? AND event_type = 'file.touched'
		ORDER BY timestamp ASC`

	rows, err := r.db.QueryContext(ctx, q, runID)
	if err != nil {
		return nil, fmt.Errorf("query: consultar activity events del run %q: %w", runID, err)
	}
	defer rows.Close()

	var results []entity.ActivityEvent
	for rows.Next() {
		var ev entity.ActivityEvent
		if err := rows.Scan(&ev.Timestamp, &ev.AgentID); err != nil {
			return nil, fmt.Errorf("query: escanear activity event: %w", err)
		}
		results = append(results, ev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query: iterar activity events: %w", err)
	}
	return results, nil
}
