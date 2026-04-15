package query

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

// ListErrorGroups returns error groups with optional filtering.
func (r *Reader) ListErrorGroups(ctx context.Context, status, search string) ([]entity.ErrorGroup, error) {
	var args []any
	clauses := []string{"1=1"}

	if status != "" {
		clauses = append(clauses, "eg.resolution_status = ?")
		args = append(args, status)
	}
	if search != "" {
		clauses = append(clauses, "(eg.title LIKE ? OR eg.normalized_msg LIKE ?)")
		s := "%" + search + "%"
		args = append(args, s, s)
	}

	q := fmt.Sprintf(`
		SELECT eg.id, eg.fingerprint, eg.title, eg.normalized_msg, eg.resolution_status,
		       eg.first_seen_at, eg.last_seen_at, eg.occurrence_count,
		       COALESCE(eg.notes, ''), COALESCE(eg.commit_link, ''),
		       eg.created_at, eg.updated_at
		FROM error_groups eg
		WHERE %s
		ORDER BY
		  CASE WHEN eg.resolution_status = 'new' THEN 0 ELSE 1 END,
		  eg.last_seen_at DESC
		LIMIT 200`, strings.Join(clauses, " AND "))

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query: listar error_groups: %w", err)
	}
	defer rows.Close()

	var result []entity.ErrorGroup
	for rows.Next() {
		eg, err := scanErrorGroup(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, eg)
	}
	return result, rows.Err()
}

// GetErrorGroup returns a single error group by ID.
func (r *Reader) GetErrorGroup(ctx context.Context, groupID string) (*entity.ErrorGroup, error) {
	const q = `
		SELECT id, fingerprint, title, normalized_msg, resolution_status,
		       first_seen_at, last_seen_at, occurrence_count,
		       COALESCE(notes, ''), COALESCE(commit_link, ''),
		       created_at, updated_at
		FROM error_groups WHERE id = ?`

	rows, err := r.db.QueryContext(ctx, q, groupID)
	if err != nil {
		return nil, fmt.Errorf("query: obtener error_groups: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	eg, err := scanErrorGroup(rows)
	if err != nil {
		return nil, err
	}
	return &eg, rows.Err()
}

// ListErrorGroupRuns returns runs linked to an error group.
func (r *Reader) ListErrorGroupRuns(ctx context.Context, groupID string) ([]entity.ErrorGroupRun, error) {
	const q = `
		SELECT run_id, agent_name, error_msg, exit_code, occurred_at
		FROM error_group_runs
		WHERE error_group_id = ?
		ORDER BY occurred_at DESC
		LIMIT 50`

	rows, err := r.db.QueryContext(ctx, q, groupID)
	if err != nil {
		return nil, fmt.Errorf("query: listar error_group_runs: %w", err)
	}
	defer rows.Close()

	var result []entity.ErrorGroupRun
	for rows.Next() {
		var egr entity.ErrorGroupRun
		if err := rows.Scan(&egr.RunID, &egr.AgentName, &egr.ErrorMsg, &egr.ExitCode, &egr.OccurredAt); err != nil {
			return nil, fmt.Errorf("query: scan error_group_runs: %w", err)
		}
		result = append(result, egr)
	}
	return result, rows.Err()
}

// ListErrorGroupHistory returns status change history for an error group.
func (r *Reader) ListErrorGroupHistory(ctx context.Context, groupID string) ([]entity.ErrorGroupHistory, error) {
	const q = `
		SELECT id, COALESCE(old_status, ''), new_status, COALESCE(note, ''), COALESCE(commit_link, ''), created_at
		FROM error_group_history
		WHERE error_group_id = ?
		ORDER BY created_at DESC
		LIMIT 20`

	rows, err := r.db.QueryContext(ctx, q, groupID)
	if err != nil {
		return nil, fmt.Errorf("query: listar error_group_history: %w", err)
	}
	defer rows.Close()

	var result []entity.ErrorGroupHistory
	for rows.Next() {
		var h entity.ErrorGroupHistory
		if err := rows.Scan(&h.ID, &h.OldStatus, &h.NewStatus, &h.Note, &h.CommitLink, &h.CreatedAt); err != nil {
			return nil, fmt.Errorf("query: scan error_group_history: %w", err)
		}
		result = append(result, h)
	}
	return result, rows.Err()
}

// GetErrorTrend returns daily error counts for the last 14 days for a group.
func (r *Reader) GetErrorTrend(ctx context.Context, groupID string) ([]entity.TrendPoint, error) {
	const q = `
		WITH RECURSIVE dates(d) AS (
			SELECT date('now', '-13 days')
			UNION ALL
			SELECT date(d, '+1 day') FROM dates WHERE d < date('now')
		)
		SELECT dates.d, COALESCE(counts.c, 0)
		FROM dates
		LEFT JOIN (
			SELECT date(occurred_at) AS day, COUNT(*) AS c
			FROM error_group_runs
			WHERE error_group_id = ?
			  AND occurred_at >= datetime('now', '-14 days')
			GROUP BY day
		) counts ON dates.d = counts.day
		ORDER BY dates.d`

	rows, err := r.db.QueryContext(ctx, q, groupID)
	if err != nil {
		return nil, fmt.Errorf("query: obtener error trend: %w", err)
	}
	defer rows.Close()

	var result []entity.TrendPoint
	for rows.Next() {
		var tp entity.TrendPoint
		if err := rows.Scan(&tp.Date, &tp.Count); err != nil {
			return nil, fmt.Errorf("query: scan error trend: %w", err)
		}
		result = append(result, tp)
	}
	return result, rows.Err()
}

// GetErrorCounts returns counts of error groups by resolution status.
func (r *Reader) GetErrorCounts(ctx context.Context) (entity.ErrorCounts, error) {
	const q = `
		SELECT
			COALESCE(SUM(CASE WHEN resolution_status = 'new' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN resolution_status = 'investigating' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN last_seen_at >= datetime('now', '-7 days') THEN 1 ELSE 0 END), 0)
		FROM error_groups`

	var ec entity.ErrorCounts
	err := r.db.QueryRowContext(ctx, q).Scan(&ec.New, &ec.Investigating, &ec.WeekTotal)
	if err != nil {
		return ec, fmt.Errorf("query: contar errores por status: %w", err)
	}
	return ec, nil
}

// --- scan helper ---

func scanErrorGroup(scanner interface{ Scan(dest ...any) error }) (entity.ErrorGroup, error) {
	var eg entity.ErrorGroup
	var firstSeen, lastSeen, created, updated string
	if err := scanner.Scan(&eg.ID, &eg.Fingerprint, &eg.Title, &eg.NormalizedMsg,
		&eg.ResolutionStatus, &firstSeen, &lastSeen, &eg.OccurrenceCount,
		&eg.Notes, &eg.CommitLink, &created, &updated); err != nil {
		return eg, fmt.Errorf("query: scan error_groups: %w", err)
	}
	eg.FirstSeenAt, _ = time.Parse(time.RFC3339Nano, firstSeen)
	eg.LastSeenAt, _ = time.Parse(time.RFC3339Nano, lastSeen)
	eg.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	eg.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
	return eg, nil
}
