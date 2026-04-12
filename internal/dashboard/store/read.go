package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// AgentRow es una fila de la tabla agents preparada para la vista de flujo.
// Los campos nullables usan punteros para preservar la diferencia entre "cero"
// y "no seteado" (p.ej. duration_ms de un agente que aún está corriendo).
type AgentRow struct {
	AgentID     string
	AgentRole   string
	Sequence    *int
	DependsOn   string // JSON array serializado como TEXT en SQLite, puede ser vacío
	Status      string
	DurationMs  *int64
	TokensTotal *int
}

const defaultListLimit = 100

// RunSummary es una fila de la tabla runs preparada para la vista de Runs.
// Los campos nullables usan punteros para preservar la diferencia entre
// "cero" y "no seteado" (p.ej. qa_score o ended_at en runs en progreso).
type RunSummary struct {
	ID            string
	TaskID        string
	TaskDesc      string
	Status        string
	Complexity    string
	Provider      string
	Project       string
	StartedAt     time.Time
	EndedAt       *time.Time
	DurationMs    *int64
	TotalTokens   int
	FilesCount    int
	AgentsCount   int
	QAScore       *float64
	ParentRunID   string
	ChildrenCount int
}

// ListRuns devuelve los runs ordenados por started_at DESC.
// limit <= 0 aplica defaultListLimit (100). offset < 0 se normaliza a 0.
// status, startDate, endDate y project son filtros opcionales (cadena vacía = sin filtro).
func (s *SQLiteStore) ListRuns(ctx context.Context, limit, offset int, status, startDate, endDate, project string) ([]RunSummary, error) {
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
			(SELECT COUNT(*) FROM runs c WHERE c.parent_run_id = runs.id) AS children_count
		FROM runs`

	// parent_run_id IS NULL siempre se aplica para listar solo runs top-level.
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

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("dashboard/store: consultar runs: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var results []RunSummary
	for rows.Next() {
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
		)

		if err := rows.Scan(
			&id, &taskID, &taskDesc, &status, &complexity, &provider, &project,
			&startedAtRaw, &endedAtRaw, &durationMs,
			&totalTokens, &filesCount, &agentsCount, &qaScore,
			&parentRunID, &childrenCount,
		); err != nil {
			return nil, fmt.Errorf("dashboard/store: escanear fila de runs: %w", err)
		}

		startedAt, err := time.Parse(time.RFC3339Nano, startedAtRaw)
		if err != nil {
			return nil, fmt.Errorf("dashboard/store: parsear started_at %q: %w", startedAtRaw, err)
		}

		r := RunSummary{
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
		}

		if endedAtRaw.Valid && endedAtRaw.String != "" {
			t, err := time.Parse(time.RFC3339Nano, endedAtRaw.String)
			if err != nil {
				return nil, fmt.Errorf("dashboard/store: parsear ended_at %q: %w", endedAtRaw.String, err)
			}
			r.EndedAt = &t
		}

		if durationMs.Valid {
			v := durationMs.Int64
			r.DurationMs = &v
		}

		if qaScore.Valid {
			v := qaScore.Float64
			r.QAScore = &v
		}

		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard/store: iterar filas de runs: %w", err)
	}

	if results == nil {
		results = []RunSummary{}
	}

	return results, nil
}

// GetRunSummary retorna el RunSummary del run identificado por runID.
// Retorna (nil, nil) si el run no existe — NO es un error.
func (s *SQLiteStore) GetRunSummary(ctx context.Context, runID string) (*RunSummary, error) {
	const q = `
		SELECT
			id, task_id, task_desc, status, complexity, provider, project,
			started_at, ended_at, duration_ms,
			total_tokens, files_touched, agents_count, qa_score,
			COALESCE(parent_run_id, ''),
			(SELECT COUNT(*) FROM runs c WHERE c.parent_run_id = runs.id) AS children_count
		FROM runs
		WHERE id = ?
		LIMIT 1`

	rows, err := s.db.QueryContext(ctx, q, runID)
	if err != nil {
		return nil, fmt.Errorf("dashboard/store: consultar run %q: %w", runID, err)
	}
	defer rows.Close() //nolint:errcheck

	if !rows.Next() {
		return nil, nil
	}

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
	)

	if err := rows.Scan(
		&id, &taskID, &taskDesc, &status, &complexity, &provider, &project,
		&startedAtRaw, &endedAtRaw, &durationMs,
		&totalTokens, &filesCount, &agentsCount, &qaScore,
		&parentRunID, &childrenCount,
	); err != nil {
		return nil, fmt.Errorf("dashboard/store: escanear fila de run: %w", err)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard/store: iterar fila de run: %w", err)
	}

	startedAt, err := time.Parse(time.RFC3339Nano, startedAtRaw)
	if err != nil {
		return nil, fmt.Errorf("dashboard/store: parsear started_at %q: %w", startedAtRaw, err)
	}

	r := RunSummary{
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
	}

	if endedAtRaw.Valid && endedAtRaw.String != "" {
		t, err := time.Parse(time.RFC3339Nano, endedAtRaw.String)
		if err != nil {
			return nil, fmt.Errorf("dashboard/store: parsear ended_at %q: %w", endedAtRaw.String, err)
		}
		r.EndedAt = &t
	}

	if durationMs.Valid {
		v := durationMs.Int64
		r.DurationMs = &v
	}

	if qaScore.Valid {
		v := qaScore.Float64
		r.QAScore = &v
	}

	return &r, nil
}

// AgentDetail es la proyección completa de un agente para la vista de detalle.
// Los campos nullables usan punteros para preservar la diferencia entre "cero" y "no seteado".
type AgentDetail struct {
	AgentID      string
	AgentRole    string
	Status       string
	StartedAt    *time.Time
	EndedAt      *time.Time
	DurationMs   *int64
	TokensInput  *int
	TokensOutput *int
	TokensTotal  *int
	ExitCode     *int
	ErrorMsg     string
	Output       string
}

// FileRow es una fila de la tabla files asociada a un agente.
type FileRow struct {
	Path      string
	Operation string
	Diff      string
}

// GetAgentDetail retorna el detalle de un agente individual (runID + agentID) junto con sus archivos.
// Retorna (nil, nil, nil) si el agente no existe — NO es un error.
// Retorna (&detail, []FileRow{}, nil) si el agente existe pero no tiene archivos.
func (s *SQLiteStore) GetAgentDetail(ctx context.Context, runID, agentID string) (*AgentDetail, []FileRow, error) {
	const agentQ = `
		SELECT agent_id, agent_role, status, started_at, ended_at, duration_ms,
		       tokens_input, tokens_output, tokens_total, exit_code, error_msg, output
		FROM agents
		WHERE run_id = ? AND agent_id = ?`

	rows, err := s.db.QueryContext(ctx, agentQ, runID, agentID)
	if err != nil {
		return nil, nil, fmt.Errorf("dashboard/store: consultar agente %q del run %q: %w", agentID, runID, err)
	}
	defer rows.Close() //nolint:errcheck

	if !rows.Next() {
		// El agente no existe — no es un error.
		return nil, nil, nil
	}

	var (
		dbAgentID      string
		dbAgentRole    string
		dbStatus       string
		dbStartedAtRaw sql.NullString
		dbEndedAtRaw   sql.NullString
		dbDurationMs   sql.NullInt64
		dbTokensInput  sql.NullInt64
		dbTokensOutput sql.NullInt64
		dbTokensTotal  sql.NullInt64
		dbExitCode     sql.NullInt64
		dbErrorMsg     sql.NullString
		dbOutput       sql.NullString
	)

	if err := rows.Scan(
		&dbAgentID, &dbAgentRole, &dbStatus,
		&dbStartedAtRaw, &dbEndedAtRaw, &dbDurationMs,
		&dbTokensInput, &dbTokensOutput, &dbTokensTotal,
		&dbExitCode, &dbErrorMsg, &dbOutput,
	); err != nil {
		return nil, nil, fmt.Errorf("dashboard/store: escanear fila de agents: %w", err)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("dashboard/store: iterar filas de agents: %w", err)
	}

	detail := AgentDetail{
		AgentID:   dbAgentID,
		AgentRole: dbAgentRole,
		Status:    dbStatus,
		ErrorMsg:  dbErrorMsg.String,
		Output:    dbOutput.String,
	}

	if dbStartedAtRaw.Valid && dbStartedAtRaw.String != "" {
		t, err := time.Parse(time.RFC3339Nano, dbStartedAtRaw.String)
		if err != nil {
			return nil, nil, fmt.Errorf("dashboard/store: parsear started_at %q: %w", dbStartedAtRaw.String, err)
		}
		detail.StartedAt = &t
	}

	if dbEndedAtRaw.Valid && dbEndedAtRaw.String != "" {
		t, err := time.Parse(time.RFC3339Nano, dbEndedAtRaw.String)
		if err != nil {
			return nil, nil, fmt.Errorf("dashboard/store: parsear ended_at %q: %w", dbEndedAtRaw.String, err)
		}
		detail.EndedAt = &t
	}

	if dbDurationMs.Valid {
		v := dbDurationMs.Int64
		detail.DurationMs = &v
	}

	if dbTokensInput.Valid {
		v := int(dbTokensInput.Int64)
		detail.TokensInput = &v
	}

	if dbTokensOutput.Valid {
		v := int(dbTokensOutput.Int64)
		detail.TokensOutput = &v
	}

	if dbTokensTotal.Valid {
		v := int(dbTokensTotal.Int64)
		detail.TokensTotal = &v
	}

	if dbExitCode.Valid {
		v := int(dbExitCode.Int64)
		detail.ExitCode = &v
	}

	// Consultar archivos del agente.
	const filesQ = `
		SELECT path, operation, COALESCE(diff, '')
		FROM files
		WHERE run_id = ? AND agent_id = ?
		ORDER BY id ASC`

	fileRows, err := s.db.QueryContext(ctx, filesQ, runID, agentID)
	if err != nil {
		return nil, nil, fmt.Errorf("dashboard/store: consultar files del agente %q: %w", agentID, err)
	}
	defer fileRows.Close() //nolint:errcheck

	files := make([]FileRow, 0)
	for fileRows.Next() {
		var path, operation, diff string
		if err := fileRows.Scan(&path, &operation, &diff); err != nil {
			return nil, nil, fmt.Errorf("dashboard/store: escanear fila de files: %w", err)
		}
		files = append(files, FileRow{Path: path, Operation: operation, Diff: diff})
	}

	if err := fileRows.Err(); err != nil {
		return nil, nil, fmt.Errorf("dashboard/store: iterar filas de files: %w", err)
	}

	return &detail, files, nil
}

// ListAgentsByRun devuelve todos los agentes de un run ordenados por sequence ASC, id ASC.
// Si el run no existe o no tiene agentes retorna nil, nil (sin error).
func (s *SQLiteStore) ListAgentsByRun(ctx context.Context, runID string) ([]AgentRow, error) {
	const q = `
		SELECT agent_id, agent_role, sequence, depends_on, status, duration_ms, tokens_total
		FROM agents
		WHERE run_id = ?
		ORDER BY sequence ASC, id ASC`

	rows, err := s.db.QueryContext(ctx, q, runID)
	if err != nil {
		return nil, fmt.Errorf("dashboard/store: consultar agents del run %q: %w", runID, err)
	}
	defer rows.Close() //nolint:errcheck

	var results []AgentRow
	for rows.Next() {
		var (
			agentID     string
			agentRole   string
			sequence    sql.NullInt64
			dependsOn   sql.NullString
			status      string
			durationMs  sql.NullInt64
			tokensTotal sql.NullInt64
		)

		if err := rows.Scan(
			&agentID, &agentRole, &sequence, &dependsOn,
			&status, &durationMs, &tokensTotal,
		); err != nil {
			return nil, fmt.Errorf("dashboard/store: escanear fila de agents: %w", err)
		}

		r := AgentRow{
			AgentID:   agentID,
			AgentRole: agentRole,
			Status:    status,
			DependsOn: dependsOn.String,
		}

		if sequence.Valid {
			v := int(sequence.Int64)
			r.Sequence = &v
		}

		if durationMs.Valid {
			v := durationMs.Int64
			r.DurationMs = &v
		}

		if tokensTotal.Valid {
			v := int(tokensTotal.Int64)
			r.TokensTotal = &v
		}

		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard/store: iterar filas de agents: %w", err)
	}

	return results, nil
}

// ListProjects returns distinct non-empty project names, ordered alphabetically.
func (s *SQLiteStore) ListProjects(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT project FROM runs WHERE project != '' ORDER BY project ASC`)
	if err != nil {
		return nil, fmt.Errorf("dashboard/store: listar proyectos: %w", err)
	}
	defer rows.Close() //nolint:errcheck

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

// ListChildRuns returns runs whose parent_run_id matches the given parentID.
// Ordered by started_at ASC (chronological within the parent).
func (s *SQLiteStore) ListChildRuns(ctx context.Context, parentRunID string) ([]RunSummary, error) {
	const q = `
		SELECT
			id, task_id, task_desc, status, complexity, provider, project,
			started_at, ended_at, duration_ms,
			total_tokens, files_touched, agents_count, qa_score,
			COALESCE(parent_run_id, ''),
			(SELECT COUNT(*) FROM runs c WHERE c.parent_run_id = runs.id) AS children_count
		FROM runs
		WHERE parent_run_id = ?
		ORDER BY started_at ASC`

	rows, err := s.db.QueryContext(ctx, q, parentRunID)
	if err != nil {
		return nil, fmt.Errorf("dashboard/store: consultar child runs de %q: %w", parentRunID, err)
	}
	defer rows.Close() //nolint:errcheck

	var results []RunSummary
	for rows.Next() {
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
			pRunID        string
			childrenCount int
		)

		if err := rows.Scan(
			&id, &taskID, &taskDesc, &status, &complexity, &provider, &project,
			&startedAtRaw, &endedAtRaw, &durationMs,
			&totalTokens, &filesCount, &agentsCount, &qaScore,
			&pRunID, &childrenCount,
		); err != nil {
			return nil, fmt.Errorf("dashboard/store: escanear fila de child run: %w", err)
		}

		startedAt, err := time.Parse(time.RFC3339Nano, startedAtRaw)
		if err != nil {
			return nil, fmt.Errorf("dashboard/store: parsear started_at %q: %w", startedAtRaw, err)
		}

		r := RunSummary{
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
			ParentRunID:   pRunID,
			ChildrenCount: childrenCount,
		}

		if endedAtRaw.Valid && endedAtRaw.String != "" {
			t, err := time.Parse(time.RFC3339Nano, endedAtRaw.String)
			if err != nil {
				return nil, fmt.Errorf("dashboard/store: parsear ended_at %q: %w", endedAtRaw.String, err)
			}
			r.EndedAt = &t
		}

		if durationMs.Valid {
			v := durationMs.Int64
			r.DurationMs = &v
		}

		if qaScore.Valid {
			v := qaScore.Float64
			r.QAScore = &v
		}

		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard/store: iterar filas de child runs: %w", err)
	}

	if results == nil {
		results = []RunSummary{}
	}

	return results, nil
}

// ListFilesByRun returns all files touched in a run, ordered by id ASC.
// Returns nil, nil if the run has no files.
func (s *SQLiteStore) ListFilesByRun(ctx context.Context, runID string) ([]FileRow, error) {
	const q = `
		SELECT path, operation, COALESCE(diff, '')
		FROM files
		WHERE run_id = ?
		ORDER BY id ASC`

	rows, err := s.db.QueryContext(ctx, q, runID)
	if err != nil {
		return nil, fmt.Errorf("dashboard/store: consultar files del run %q: %w", runID, err)
	}
	defer rows.Close() //nolint:errcheck

	var results []FileRow
	for rows.Next() {
		var f FileRow
		if err := rows.Scan(&f.Path, &f.Operation, &f.Diff); err != nil {
			return nil, fmt.Errorf("dashboard/store: escanear fila de files: %w", err)
		}
		results = append(results, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard/store: iterar filas de files: %w", err)
	}
	return results, nil
}
