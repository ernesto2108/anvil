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
	ID          string
	TaskID      string
	TaskDesc    string
	Status      string
	Complexity  string
	Provider    string
	StartedAt   time.Time
	EndedAt     *time.Time
	DurationMs  *int64
	TotalTokens int
	FilesCount  int
	AgentsCount int
	QAScore     *float64
}

// ListRuns devuelve los runs ordenados por started_at DESC.
// limit <= 0 aplica defaultListLimit (100). offset < 0 se normaliza a 0.
func (s *SQLiteStore) ListRuns(ctx context.Context, limit, offset int) ([]RunSummary, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if offset < 0 {
		offset = 0
	}

	const q = `
		SELECT
			id, task_id, task_desc, status, complexity, provider,
			started_at, ended_at, duration_ms,
			total_tokens, files_touched, agents_count, qa_score
		FROM runs
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?`

	rows, err := s.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("dashboard/store: consultar runs: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var results []RunSummary
	for rows.Next() {
		var (
			id           string
			taskID       string
			taskDesc     sql.NullString
			status       string
			complexity   sql.NullString
			provider     sql.NullString
			startedAtRaw string
			endedAtRaw   sql.NullString
			durationMs   sql.NullInt64
			totalTokens  int
			filesCount   int
			agentsCount  int
			qaScore      sql.NullFloat64
		)

		if err := rows.Scan(
			&id, &taskID, &taskDesc, &status, &complexity, &provider,
			&startedAtRaw, &endedAtRaw, &durationMs,
			&totalTokens, &filesCount, &agentsCount, &qaScore,
		); err != nil {
			return nil, fmt.Errorf("dashboard/store: escanear fila de runs: %w", err)
		}

		startedAt, err := time.Parse(time.RFC3339Nano, startedAtRaw)
		if err != nil {
			return nil, fmt.Errorf("dashboard/store: parsear started_at %q: %w", startedAtRaw, err)
		}

		r := RunSummary{
			ID:          id,
			TaskID:      taskID,
			TaskDesc:    taskDesc.String,
			Status:      status,
			Complexity:  complexity.String,
			Provider:    provider.String,
			StartedAt:   startedAt,
			TotalTokens: totalTokens,
			FilesCount:  filesCount,
			AgentsCount: agentsCount,
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
