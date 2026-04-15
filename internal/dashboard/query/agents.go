package query

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

// ListAgentsByRun returns all agents for a run ordered by sequence ASC, id ASC.
func (r *Reader) ListAgentsByRun(ctx context.Context, runID string) ([]entity.AgentRow, error) {
	const q = `
		SELECT agent_id, agent_role, sequence, depends_on, status,
		       started_at, ended_at, duration_ms, tokens_total
		FROM agents
		WHERE run_id = ?
		ORDER BY sequence ASC, id ASC`

	rows, err := r.db.QueryContext(ctx, q, runID)
	if err != nil {
		return nil, fmt.Errorf("query: consultar agents del run %q: %w", runID, err)
	}
	defer rows.Close()

	var results []entity.AgentRow
	for rows.Next() {
		var (
			agentID      string
			agentRole    string
			sequence     sql.NullInt64
			dependsOn    sql.NullString
			status       string
			startedAtRaw sql.NullString
			endedAtRaw   sql.NullString
			durationMs   sql.NullInt64
			tokensTotal  sql.NullInt64
		)

		if err := rows.Scan(
			&agentID, &agentRole, &sequence, &dependsOn,
			&status, &startedAtRaw, &endedAtRaw, &durationMs, &tokensTotal,
		); err != nil {
			return nil, fmt.Errorf("query: escanear fila de agents: %w", err)
		}

		ar := entity.AgentRow{
			AgentID:   agentID,
			AgentRole: agentRole,
			Status:    status,
			DependsOn: dependsOn.String,
		}

		if sequence.Valid {
			v := int(sequence.Int64)
			ar.Sequence = &v
		}
		if startedAtRaw.Valid && startedAtRaw.String != "" {
			t, err := time.Parse(time.RFC3339Nano, startedAtRaw.String)
			if err == nil {
				ar.StartedAt = &t
			}
		}
		if endedAtRaw.Valid && endedAtRaw.String != "" {
			t, err := time.Parse(time.RFC3339Nano, endedAtRaw.String)
			if err == nil {
				ar.EndedAt = &t
			}
		}
		if durationMs.Valid {
			v := durationMs.Int64
			ar.DurationMs = &v
		}
		if tokensTotal.Valid {
			v := int(tokensTotal.Int64)
			ar.TokensTotal = &v
		}

		results = append(results, ar)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query: iterar filas de agents: %w", err)
	}
	return results, nil
}

// GetAgentDetail returns the full detail of a single agent within a run.
// Returns (nil, nil, nil) if the agent does not exist.
func (r *Reader) GetAgentDetail(ctx context.Context, runID, agentID string) (*entity.AgentDetail, []entity.FileRecord, error) {
	const agentQ = `
		SELECT agent_id, agent_role, status, started_at, ended_at, duration_ms,
		       tokens_input, tokens_output, tokens_total, exit_code, error_msg, output
		FROM agents
		WHERE run_id = ? AND agent_id = ?`

	rows, err := r.db.QueryContext(ctx, agentQ, runID, agentID)
	if err != nil {
		return nil, nil, fmt.Errorf("query: consultar agente %q del run %q: %w", agentID, runID, err)
	}
	defer rows.Close()

	if !rows.Next() {
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
		return nil, nil, fmt.Errorf("query: escanear fila de agents: %w", err)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("query: iterar filas de agents: %w", err)
	}
	rows.Close()

	detail := entity.AgentDetail{
		AgentID:   dbAgentID,
		AgentRole: dbAgentRole,
		Status:    dbStatus,
		ErrorMsg:  dbErrorMsg.String,
		Output:    dbOutput.String,
	}

	if dbStartedAtRaw.Valid && dbStartedAtRaw.String != "" {
		t, err := time.Parse(time.RFC3339Nano, dbStartedAtRaw.String)
		if err != nil {
			return nil, nil, fmt.Errorf("query: parsear started_at %q: %w", dbStartedAtRaw.String, err)
		}
		detail.StartedAt = &t
	}
	if dbEndedAtRaw.Valid && dbEndedAtRaw.String != "" {
		t, err := time.Parse(time.RFC3339Nano, dbEndedAtRaw.String)
		if err != nil {
			return nil, nil, fmt.Errorf("query: parsear ended_at %q: %w", dbEndedAtRaw.String, err)
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

	// Query agent files.
	const filesQ = `
		SELECT path, operation, COALESCE(diff, ''), agent_id
		FROM files
		WHERE run_id = ? AND agent_id = ?
		ORDER BY id ASC`

	fileRows, err := r.db.QueryContext(ctx, filesQ, runID, agentID)
	if err != nil {
		return nil, nil, fmt.Errorf("query: consultar files del agente %q: %w", agentID, err)
	}
	defer fileRows.Close()

	files := make([]entity.FileRecord, 0)
	for fileRows.Next() {
		var f entity.FileRecord
		if err := fileRows.Scan(&f.Path, &f.Operation, &f.Diff, &f.AgentID); err != nil {
			return nil, nil, fmt.Errorf("query: escanear fila de files: %w", err)
		}
		files = append(files, f)
	}
	if err := fileRows.Err(); err != nil {
		return nil, nil, fmt.Errorf("query: iterar filas de files: %w", err)
	}

	return &detail, files, nil
}
