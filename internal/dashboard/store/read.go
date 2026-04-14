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
	StartedAt   *time.Time
	EndedAt     *time.Time
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
	Branch        string
	ErrorReason   string
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
			(SELECT COUNT(*) FROM runs c WHERE c.parent_run_id = runs.id) AS children_count,
			COALESCE(branch, ''),
			COALESCE(error_reason, '')
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
			branch        string
			errorReason   string
		)

		if err := rows.Scan(
			&id, &taskID, &taskDesc, &status, &complexity, &provider, &project,
			&startedAtRaw, &endedAtRaw, &durationMs,
			&totalTokens, &filesCount, &agentsCount, &qaScore,
			&parentRunID, &childrenCount, &branch, &errorReason,
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
			Branch:        branch,
			ErrorReason:   errorReason,
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
			(SELECT COUNT(*) FROM runs c WHERE c.parent_run_id = runs.id) AS children_count,
			COALESCE(branch, ''),
			COALESCE(error_reason, '')
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
		branch        string
		errorReason   string
	)

	if err := rows.Scan(
		&id, &taskID, &taskDesc, &status, &complexity, &provider, &project,
		&startedAtRaw, &endedAtRaw, &durationMs,
		&totalTokens, &filesCount, &agentsCount, &qaScore,
		&parentRunID, &childrenCount, &branch, &errorReason,
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
		Branch:        branch,
		ErrorReason:   errorReason,
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
	AgentID   string // "" si fue editado directamente por la sesión principal
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

	// Close rows BEFORE the second query to release the connection.
	// With MaxOpenConns(1), holding rows open while opening fileRows deadlocks.
	rows.Close()

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
		SELECT path, operation, COALESCE(diff, ''), agent_id
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
		var f FileRow
		if err := fileRows.Scan(&f.Path, &f.Operation, &f.Diff, &f.AgentID); err != nil {
			return nil, nil, fmt.Errorf("dashboard/store: escanear fila de files: %w", err)
		}
		files = append(files, f)
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
		SELECT agent_id, agent_role, sequence, depends_on, status,
		       started_at, ended_at, duration_ms, tokens_total
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
			agentID       string
			agentRole     string
			sequence      sql.NullInt64
			dependsOn     sql.NullString
			status        string
			startedAtRaw  sql.NullString
			endedAtRaw    sql.NullString
			durationMs    sql.NullInt64
			tokensTotal   sql.NullInt64
		)

		if err := rows.Scan(
			&agentID, &agentRole, &sequence, &dependsOn,
			&status, &startedAtRaw, &endedAtRaw, &durationMs, &tokensTotal,
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

		if startedAtRaw.Valid && startedAtRaw.String != "" {
			t, err := time.Parse(time.RFC3339Nano, startedAtRaw.String)
			if err == nil {
				r.StartedAt = &t
			}
		}

		if endedAtRaw.Valid && endedAtRaw.String != "" {
			t, err := time.Parse(time.RFC3339Nano, endedAtRaw.String)
			if err == nil {
				r.EndedAt = &t
			}
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
			(SELECT COUNT(*) FROM runs c WHERE c.parent_run_id = runs.id) AS children_count,
			COALESCE(branch, ''),
			COALESCE(error_reason, '')
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
			branch        string
			errorReason   string
		)

		if err := rows.Scan(
			&id, &taskID, &taskDesc, &status, &complexity, &provider, &project,
			&startedAtRaw, &endedAtRaw, &durationMs,
			&totalTokens, &filesCount, &agentsCount, &qaScore,
			&pRunID, &childrenCount, &branch, &errorReason,
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
			Branch:        branch,
			ErrorReason:   errorReason,
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
		SELECT path, operation, COALESCE(diff, ''), agent_id
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
		if err := rows.Scan(&f.Path, &f.Operation, &f.Diff, &f.AgentID); err != nil {
			return nil, fmt.Errorf("dashboard/store: escanear fila de files: %w", err)
		}
		results = append(results, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard/store: iterar filas de files: %w", err)
	}
	return results, nil
}

// ActivityEvent is a timestamp + agent attribution for a file edit event.
type ActivityEvent struct {
	Timestamp string
	AgentID   string // "" = main session (Claude direct)
}

// ListActivityEvents returns file.touched events with timestamps and agent attribution.
// Used to render activity markers on the session trace timeline.
func (s *SQLiteStore) ListActivityEvents(ctx context.Context, runID string) ([]ActivityEvent, error) {
	// Extract agent_id from the JSON payload using SQLite json_extract.
	const q = `
		SELECT timestamp, COALESCE(json_extract(payload, '$.agent_id'), '')
		FROM events
		WHERE run_id = ? AND event_type = 'file.touched'
		ORDER BY timestamp ASC`

	rows, err := s.db.QueryContext(ctx, q, runID)
	if err != nil {
		return nil, fmt.Errorf("dashboard/store: consultar activity events del run %q: %w", runID, err)
	}
	defer rows.Close() //nolint:errcheck

	var results []ActivityEvent
	for rows.Next() {
		var ev ActivityEvent
		if err := rows.Scan(&ev.Timestamp, &ev.AgentID); err != nil {
			return nil, fmt.Errorf("dashboard/store: escanear activity event: %w", err)
		}
		results = append(results, ev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard/store: iterar activity events: %w", err)
	}
	return results, nil
}

// --- New hook-driven queries -------------------------------------------------

// ToolUseSummary is a count of tool invocations grouped by tool name.
type ToolUseSummary struct {
	ToolName string
	Count    int
}

// ListToolUsageByRun returns tool usage counts for a run, ordered by count DESC.
func (s *SQLiteStore) ListToolUsageByRun(ctx context.Context, runID string) ([]ToolUseSummary, error) {
	const q = `SELECT tool_name, COUNT(*) AS cnt FROM tool_uses WHERE run_id = ? GROUP BY tool_name ORDER BY cnt DESC`

	rows, err := s.db.QueryContext(ctx, q, runID)
	if err != nil {
		return nil, fmt.Errorf("dashboard/store: consultar tool_uses del run %q: %w", runID, err)
	}
	defer rows.Close() //nolint:errcheck

	var results []ToolUseSummary
	for rows.Next() {
		var ts ToolUseSummary
		if err := rows.Scan(&ts.ToolName, &ts.Count); err != nil {
			return nil, fmt.Errorf("dashboard/store: escanear tool_uses: %w", err)
		}
		results = append(results, ts)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard/store: iterar tool_uses: %w", err)
	}
	return results, nil
}

// ToolUseDetail is a single tool invocation with its input.
type ToolUseDetail struct {
	ToolName  string
	ToolInput string
	AgentID   string
	Timestamp string
}

// ListToolUseDetailsByRun returns individual tool invocations for a run,
// filtered to tools that have a non-empty input (mainly Bash).
func (s *SQLiteStore) ListToolUseDetailsByRun(ctx context.Context, runID string) ([]ToolUseDetail, error) {
	const q = `SELECT tool_name, tool_input, agent_id, timestamp
	           FROM tool_uses
	           WHERE run_id = ? AND tool_input != ''
	           ORDER BY timestamp ASC`

	rows, err := s.db.QueryContext(ctx, q, runID)
	if err != nil {
		return nil, fmt.Errorf("dashboard/store: consultar tool_use details del run %q: %w", runID, err)
	}
	defer rows.Close() //nolint:errcheck

	var results []ToolUseDetail
	for rows.Next() {
		var d ToolUseDetail
		if err := rows.Scan(&d.ToolName, &d.ToolInput, &d.AgentID, &d.Timestamp); err != nil {
			return nil, fmt.Errorf("dashboard/store: escanear tool_use detail: %w", err)
		}
		results = append(results, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard/store: iterar tool_use details: %w", err)
	}
	return results, nil
}

// TotalToolUsesByRun returns the total count of tool invocations for a run.
func (s *SQLiteStore) TotalToolUsesByRun(ctx context.Context, runID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tool_uses WHERE run_id = ?`, runID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("dashboard/store: contar tool_uses: %w", err)
	}
	return count, nil
}

// TaskRow is a row from the tasks table.
type TaskRow struct {
	TaskID      string
	Title       string
	Status      string
	CreatedAt   string
	CompletedAt string
}

// ListTasksByRun returns all tasks for a run, ordered by creation time.
func (s *SQLiteStore) ListTasksByRun(ctx context.Context, runID string) ([]TaskRow, error) {
	const q = `SELECT task_id, title, status, created_at, COALESCE(completed_at, '') FROM tasks WHERE run_id = ? ORDER BY id ASC`

	rows, err := s.db.QueryContext(ctx, q, runID)
	if err != nil {
		return nil, fmt.Errorf("dashboard/store: consultar tasks del run %q: %w", runID, err)
	}
	defer rows.Close() //nolint:errcheck

	var results []TaskRow
	for rows.Next() {
		var t TaskRow
		if err := rows.Scan(&t.TaskID, &t.Title, &t.Status, &t.CreatedAt, &t.CompletedAt); err != nil {
			return nil, fmt.Errorf("dashboard/store: escanear task: %w", err)
		}
		results = append(results, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard/store: iterar tasks: %w", err)
	}
	return results, nil
}

// CountCompactions returns the number of context.compacted events for a run.
func (s *SQLiteStore) CountCompactions(ctx context.Context, runID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM events WHERE run_id = ? AND event_type = 'context.compacted'`, runID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("dashboard/store: contar compactaciones: %w", err)
	}
	return count, nil
}

// CountPermissionDenied returns the number of permission.denied events for a run.
func (s *SQLiteStore) CountPermissionDenied(ctx context.Context, runID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM events WHERE run_id = ? AND event_type = 'permission.denied'`, runID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("dashboard/store: contar permission denied: %w", err)
	}
	return count, nil
}

// PromptRow es una fila de la tabla prompts para un run.
type PromptRow struct {
	Sequence  int
	Prompt    string
	Timestamp string
}

// TurnStatsRow contiene estadísticas de actividad para un turn (periodo entre prompts consecutivos).
type TurnStatsRow struct {
	TurnNumber    int
	Prompt        string
	Timestamp     string // inicio del turn (= timestamp del prompt)
	EndTimestamp  string // inicio del siguiente turn, o ended_at del run
	FilesCount    int
	ToolUsesCount int
	AgentsCount   int
}

// ListPromptsByRun retorna todos los prompts de un run ordenados por sequence ASC.
// Retorna nil, nil si no hay prompts (consistente con ListAgentsByRun).
func (s *SQLiteStore) ListPromptsByRun(ctx context.Context, runID string) ([]PromptRow, error) {
	const q = `SELECT sequence, prompt, timestamp FROM prompts WHERE run_id = ? ORDER BY sequence ASC`

	rows, err := s.db.QueryContext(ctx, q, runID)
	if err != nil {
		return nil, fmt.Errorf("dashboard/store: consultar prompts del run %q: %w", runID, err)
	}
	defer rows.Close() //nolint:errcheck

	var results []PromptRow
	for rows.Next() {
		var r PromptRow
		if err := rows.Scan(&r.Sequence, &r.Prompt, &r.Timestamp); err != nil {
			return nil, fmt.Errorf("dashboard/store: escanear fila de prompts: %w", err)
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard/store: iterar filas de prompts: %w", err)
	}
	return results, nil
}

// GetTurnStats retorna estadísticas de actividad por turn para un run.
// Un turn es el periodo entre prompt N y prompt N+1.
// El último turn se extiende hasta ended_at del run (o datetime('now') si sigue activo).
// Retorna nil, nil si no hay prompts.
func (s *SQLiteStore) GetTurnStats(ctx context.Context, runID string) ([]TurnStatsRow, error) {
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

	rows, err := s.db.QueryContext(ctx, q, runID, runID, runID, runID, runID)
	if err != nil {
		return nil, fmt.Errorf("dashboard/store: consultar turn stats del run %q: %w", runID, err)
	}
	defer rows.Close() //nolint:errcheck

	var results []TurnStatsRow
	for rows.Next() {
		var r TurnStatsRow
		if err := rows.Scan(
			&r.TurnNumber, &r.Prompt, &r.Timestamp, &r.EndTimestamp,
			&r.FilesCount, &r.ToolUsesCount, &r.AgentsCount,
		); err != nil {
			return nil, fmt.Errorf("dashboard/store: escanear fila de turn stats: %w", err)
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard/store: iterar filas de turn stats: %w", err)
	}
	return results, nil
}
