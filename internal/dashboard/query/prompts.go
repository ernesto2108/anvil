package query

import (
	"context"
	"fmt"

	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

// ListPromptsByRun returns all prompts for a run ordered by sequence ASC.
func (r *Reader) ListPromptsByRun(ctx context.Context, runID string) ([]entity.Prompt, error) {
	const q = `SELECT sequence, prompt, timestamp, output FROM prompts WHERE run_id = ? ORDER BY sequence ASC`

	rows, err := r.db.QueryContext(ctx, q, runID)
	if err != nil {
		return nil, fmt.Errorf("query: consultar prompts del run %q: %w", runID, err)
	}
	defer rows.Close()

	var results []entity.Prompt
	for rows.Next() {
		var p entity.Prompt
		if err := rows.Scan(&p.Sequence, &p.Prompt, &p.Timestamp, &p.Output); err != nil {
			return nil, fmt.Errorf("query: escanear fila de prompts: %w", err)
		}
		results = append(results, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query: iterar filas de prompts: %w", err)
	}
	return results, nil
}

// GetTurnStats returns activity statistics per turn for a run.
func (r *Reader) GetTurnStats(ctx context.Context, runID string) ([]entity.TurnStats, error) {
	const q = `
		WITH turn_ranges AS (
			SELECT
				p.sequence AS turn_number,
				p.prompt,
				p.timestamp AS start_ts,
				COALESCE(
					LEAD(p.timestamp) OVER (ORDER BY p.sequence),
					(SELECT COALESCE(ended_at, strftime('%Y-%m-%dT%H:%M:%fZ', 'now')) FROM runs WHERE id = ?)
				) AS end_ts
			FROM prompts p
			WHERE p.run_id = ?
		)
		SELECT
			tr.turn_number,
			tr.prompt,
			tr.start_ts,
			tr.end_ts,
			(
				SELECT COUNT(DISTINCT json_extract(e.payload, '$.path'))
				FROM events e
				WHERE e.run_id = ?
				  AND e.event_type = 'file.touched'
				  AND e.timestamp >= tr.start_ts
				  AND e.timestamp < tr.end_ts
			) AS files_count,
			(
				SELECT COUNT(*)
				FROM tool_uses tu
				WHERE tu.run_id = ?
				  AND tu.timestamp >= tr.start_ts
				  AND tu.timestamp < tr.end_ts
			) AS tool_uses_count,
			(
				SELECT COUNT(*)
				FROM agents ag
				WHERE ag.run_id = ?
				  AND ag.started_at >= tr.start_ts
				  AND ag.started_at < tr.end_ts
			) AS agents_count
		FROM turn_ranges tr
		ORDER BY tr.turn_number ASC`

	rows, err := r.db.QueryContext(ctx, q, runID, runID, runID, runID, runID)
	if err != nil {
		return nil, fmt.Errorf("query: consultar turn stats del run %q: %w", runID, err)
	}
	defer rows.Close()

	var results []entity.TurnStats
	for rows.Next() {
		var ts entity.TurnStats
		if err := rows.Scan(
			&ts.TurnNumber, &ts.Prompt, &ts.Timestamp, &ts.EndTimestamp,
			&ts.FilesCount, &ts.ToolUsesCount, &ts.AgentsCount,
		); err != nil {
			return nil, fmt.Errorf("query: escanear fila de turn stats: %w", err)
		}
		results = append(results, ts)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query: iterar filas de turn stats: %w", err)
	}
	return results, nil
}
