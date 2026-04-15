package query

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

// ListRuns returns runs ordered by started_at DESC.
// limit <= 0 applies defaultListLimit (100). offset < 0 normalizes to 0.
// status, startDate, endDate and project are optional filters (empty string = no filter).
func (r *Reader) ListRuns(ctx context.Context, limit, offset int, status, startDate, endDate, project string) ([]entity.RunSummary, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if offset < 0 {
		offset = 0
	}

	q := `
		SELECT
			id, task_id, task_desc, status, complexity, provider, project,
			started_at, ended_at, duration_ms,
			total_tokens, files_touched, agents_count, qa_score,
			COALESCE(parent_run_id, ''),
			(SELECT COUNT(*) FROM runs c WHERE c.parent_run_id = runs.id) AS children_count,
			COALESCE(branch, ''),
			COALESCE(error_reason, '')
		FROM runs`

	conditions := []string{"parent_run_id IS NULL"}
	var args []any

	if status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}
	if startDate != "" {
		conditions = append(conditions, "started_at >= ?")
		args = append(args, startDate)
	}
	if endDate != "" {
		conditions = append(conditions, "started_at <= ?")
		args = append(args, endDate)
	}
	if project != "" {
		conditions = append(conditions, "project = ?")
		args = append(args, project)
	}

	q += " WHERE " + conditions[0]
	for _, c := range conditions[1:] {
		q += " AND " + c
	}

	q += " ORDER BY started_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query: consultar runs: %w", err)
	}
	defer rows.Close()

	results, err := scanRunSummaries(rows)
	if err != nil {
		return nil, err
	}
	if results == nil {
		results = []entity.RunSummary{}
	}
	return results, nil
}

// GetRunSummary returns the RunSummary for the given runID.
// Returns (nil, nil) if the run does not exist.
func (r *Reader) GetRunSummary(ctx context.Context, runID string) (*entity.RunSummary, error) {
	const q = `
		SELECT
			id, task_id, task_desc, status, complexity, provider, project,
			started_at, ended_at, duration_ms,
			total_tokens, files_touched, agents_count, qa_score,
			COALESCE(parent_run_id, ''),
			(SELECT COUNT(*) FROM runs c WHERE c.parent_run_id = runs.id) AS children_count,
			COALESCE(branch, ''),
			COALESCE(error_reason, '')
		FROM runs
		WHERE id = ?
		LIMIT 1`

	rows, err := r.db.QueryContext(ctx, q, runID)
	if err != nil {
		return nil, fmt.Errorf("query: consultar run %q: %w", runID, err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	rs, err := scanOneRunSummary(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query: iterar fila de run: %w", err)
	}
	return &rs, nil
}

// ListChildRuns returns runs whose parent_run_id matches the given parentID.
func (r *Reader) ListChildRuns(ctx context.Context, parentRunID string) ([]entity.RunSummary, error) {
	const q = `
		SELECT
			id, task_id, task_desc, status, complexity, provider, project,
			started_at, ended_at, duration_ms,
			total_tokens, files_touched, agents_count, qa_score,
			COALESCE(parent_run_id, ''),
			(SELECT COUNT(*) FROM runs c WHERE c.parent_run_id = runs.id) AS children_count,
			COALESCE(branch, ''),
			COALESCE(error_reason, '')
		FROM runs
		WHERE parent_run_id = ?
		ORDER BY started_at ASC`

	rows, err := r.db.QueryContext(ctx, q, parentRunID)
	if err != nil {
		return nil, fmt.Errorf("query: consultar child runs de %q: %w", parentRunID, err)
	}
	defer rows.Close()

	results, err := scanRunSummaries(rows)
	if err != nil {
		return nil, err
	}
	if results == nil {
		results = []entity.RunSummary{}
	}
	return results, nil
}

// ListProjects returns distinct non-empty project names, ordered alphabetically.
func (r *Reader) ListProjects(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT DISTINCT project FROM runs WHERE project != '' ORDER BY project ASC`)
	if err != nil {
		return nil, fmt.Errorf("query: listar proyectos: %w", err)
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		results = append(results, p)
	}
	return results, rows.Err()
}

// --- scan helpers ---

func scanRunSummaries(rows *sql.Rows) ([]entity.RunSummary, error) {
	var results []entity.RunSummary
	for rows.Next() {
		rs, err := scanOneRunSummary(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, rs)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query: iterar filas de runs: %w", err)
	}
	return results, nil
}

func scanOneRunSummary(scanner interface{ Scan(dest ...any) error }) (entity.RunSummary, error) {
	var (
		id            string
		taskID        string
		taskDesc      sql.NullString
		status        string
		complexity    sql.NullString
		provider      sql.NullString
		project       sql.NullString
		startedAtRaw  string
		endedAtRaw    sql.NullString
		durationMs    sql.NullInt64
		totalTokens   int
		filesCount    int
		agentsCount   int
		qaScore       sql.NullFloat64
		parentRunID   string
		childrenCount int
		branch        string
		errorReason   string
	)

	if err := scanner.Scan(
		&id, &taskID, &taskDesc, &status, &complexity, &provider, &project,
		&startedAtRaw, &endedAtRaw, &durationMs,
		&totalTokens, &filesCount, &agentsCount, &qaScore,
		&parentRunID, &childrenCount, &branch, &errorReason,
	); err != nil {
		return entity.RunSummary{}, fmt.Errorf("query: escanear fila de runs: %w", err)
	}

	startedAt, err := time.Parse(time.RFC3339Nano, startedAtRaw)
	if err != nil {
		return entity.RunSummary{}, fmt.Errorf("query: parsear started_at %q: %w", startedAtRaw, err)
	}

	rs := entity.RunSummary{
		ID:            id,
		TaskID:        taskID,
		TaskDesc:      taskDesc.String,
		Status:        status,
		Complexity:    complexity.String,
		Provider:      provider.String,
		Project:       project.String,
		StartedAt:     startedAt,
		TotalTokens:   totalTokens,
		FilesCount:    filesCount,
		AgentsCount:   agentsCount,
		ParentRunID:   parentRunID,
		ChildrenCount: childrenCount,
		Branch:        branch,
		ErrorReason:   errorReason,
	}

	if endedAtRaw.Valid && endedAtRaw.String != "" {
		t, err := time.Parse(time.RFC3339Nano, endedAtRaw.String)
		if err != nil {
			return entity.RunSummary{}, fmt.Errorf("query: parsear ended_at %q: %w", endedAtRaw.String, err)
		}
		rs.EndedAt = &t
	}

	if durationMs.Valid {
		v := durationMs.Int64
		rs.DurationMs = &v
	}

	if qaScore.Valid {
		v := qaScore.Float64
		rs.QAScore = &v
	}

	return rs, nil
}
