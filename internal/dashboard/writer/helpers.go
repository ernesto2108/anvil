package writer

import (
	"database/sql"
	"fmt"
	"strings"
)

// expectRowsAffected returns an error when an UPDATE affects zero rows.
func expectRowsAffected(res sql.Result, table, id string) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("dashboard/writer: obtener rows affected en %s: %w", table, err)
	}
	if n == 0 {
		return fmt.Errorf("dashboard/writer: %s no encontrado en %s", id, table)
	}
	return nil
}

// ResolveRunBySession returns the run_id associated with the given session_id.
// Returns ("", nil) if no run exists for that session.
func (w *EventWriter) ResolveRunBySession(sessionID string) (string, error) {
	var runID string
	err := w.db.QueryRow(
		`SELECT id FROM runs WHERE session_id = ? LIMIT 1`, sessionID,
	).Scan(&runID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("dashboard/writer: resolver run por session_id: %w", err)
	}
	return runID, nil
}

// UpdateTaskDesc sets task_desc for a run only if it is currently empty or NULL.
func (w *EventWriter) UpdateTaskDesc(runID, desc string) error {
	const q = `UPDATE runs SET task_desc = ? WHERE id = ? AND (task_desc IS NULL OR task_desc = '')`
	_, err := w.db.Exec(q, desc, runID)
	if err != nil {
		return fmt.Errorf("dashboard/writer: actualizar task_desc: %w", err)
	}
	return nil
}

// ComputeRunTotals calculates files_touched, agents_count, and duration_ms
// from the DB and updates the run row. Called when closing a run via SessionEnd.
func (w *EventWriter) ComputeRunTotals(runID string) error {
	const q = `
		UPDATE runs
		SET files_touched = (SELECT COUNT(DISTINCT path) FROM files WHERE run_id = ?),
		    agents_count  = (SELECT COUNT(*) FROM agents WHERE run_id = ?),
		    duration_ms   = CASE
		        WHEN started_at IS NOT NULL AND ended_at IS NOT NULL
		        THEN CAST((julianday(ended_at) - julianday(started_at)) * 86400000 AS INTEGER)
		        ELSE duration_ms
		    END
		WHERE id = ?`
	_, err := w.db.Exec(q, runID, runID, runID)
	if err != nil {
		return fmt.Errorf("dashboard/writer: calcular totales de run: %w", err)
	}
	return nil
}

// CleanupStaleRuns marks runs stuck in 'running' status as 'abandoned'
// if they have no recent activity (last event) for more than the given
// minutes threshold. Called at dashboard startup to clean up orphaned sessions.
func (w *EventWriter) CleanupStaleRuns(staleMinutes int) (int64, error) {
	const q = `
		UPDATE runs
		SET status      = 'abandoned',
		    ended_at    = strftime('%Y-%m-%dT%H:%M:%fZ', 'now'),
		    duration_ms = CAST((julianday('now') - julianday(started_at)) * 86400000 AS INTEGER)
		WHERE status = 'running'
		  AND COALESCE(
		        (SELECT MAX(timestamp) FROM events WHERE events.run_id = runs.id),
		        started_at
		      ) < strftime('%Y-%m-%dT%H:%M:%fZ', 'now', ? || ' minutes')`
	res, err := w.db.Exec(q, fmt.Sprintf("-%d", staleMinutes))
	if err != nil {
		return 0, fmt.Errorf("dashboard/writer: limpiar runs huérfanos: %w", err)
	}
	return res.RowsAffected()
}

// BackfillProjects sets the project column for runs that have project = ''
// by deriving it from the most common file path prefix in the files table.
func (w *EventWriter) BackfillProjects() (int64, error) {
	rows, err := w.db.Query(`
		SELECT DISTINCT r.id, f.path
		FROM runs r
		JOIN files f ON f.run_id = r.id
		WHERE r.project = '' AND f.path LIKE '/%'
		ORDER BY r.id`)
	if err != nil {
		return 0, fmt.Errorf("dashboard/writer: backfill projects query: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	// Collect first file path per run
	runProject := make(map[string]string)
	for rows.Next() {
		var runID, path string
		if err := rows.Scan(&runID, &path); err != nil {
			return 0, err
		}
		if _, ok := runProject[runID]; ok {
			continue // already have one
		}
		runProject[runID] = extractProjectFromPath(path)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	// Close rows BEFORE the UPDATE loop to release the connection.
	// With MaxOpenConns(1), holding rows open while calling Exec deadlocks.
	rows.Close()

	var updated int64
	for runID, project := range runProject {
		if project == "" {
			continue
		}
		res, err := w.db.Exec(`UPDATE runs SET project = ? WHERE id = ? AND project = ''`, project, runID)
		if err != nil {
			return updated, err
		}
		n, _ := res.RowsAffected()
		updated += n
	}
	return updated, nil
}

// extractProjectFromPath finds the project directory name from a file path.
// Heuristic: look for a "projects" or "repos" parent, take the next segment.
// Fallback: take the 4th path segment (usually /Users/x/projects/NAME/...).
func extractProjectFromPath(path string) string {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		if (p == "projects" || p == "repos" || p == "src" || p == "workspace") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	if len(parts) >= 5 {
		return parts[4]
	}
	return ""
}

// AgentStartedAt returns the started_at timestamp for a given agent.
// Returns zero time if not found.
func (w *EventWriter) AgentStartedAt(runID, agentID string) (string, bool) {
	var startedAt string
	err := w.db.QueryRow(
		`SELECT started_at FROM agents WHERE run_id = ? AND agent_id = ? LIMIT 1`,
		runID, agentID,
	).Scan(&startedAt)
	if err != nil {
		return "", false
	}
	return startedAt, true
}

// AppendAgentOutput appends text to the output column of an agent row,
// separating turns with a blank line. Used by the Stop hook to accumulate
// Claude's responses across multiple turns in a direct session.
func (w *EventWriter) AppendAgentOutput(runID, agentID, text string) error {
	const q = `
		UPDATE agents
		SET output = CASE
			WHEN output IS NULL OR output = '' THEN ?
			ELSE output || char(10) || char(10) || '---' || char(10) || char(10) || ?
		END
		WHERE run_id = ? AND agent_id = ?`
	_, err := w.db.Exec(q, text, text, runID, agentID)
	if err != nil {
		return fmt.Errorf("dashboard/writer: append agent output: %w", err)
	}
	return nil
}

// ActiveAgentID returns the agent_id of the most recently started agent
// with status "running" for the given run. Returns "" if none found.
func (w *EventWriter) ActiveAgentID(runID string) string {
	var agentID string
	err := w.db.QueryRow(
		`SELECT agent_id FROM agents WHERE run_id = ? AND status = 'running' ORDER BY started_at DESC LIMIT 1`,
		runID,
	).Scan(&agentID)
	if err != nil {
		return ""
	}
	return agentID
}
