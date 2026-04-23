package writer

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ernesto2108/anvil/internal/instrumentation"
)

const timestampFmt = "2006-01-02T15:04:05.999999999Z07:00"

func handleRunStart(tx *sql.Tx, ev instrumentation.Event) error {
	p, err := marshalPayload[instrumentation.RunStartPayload](ev)
	if err != nil {
		return err
	}

	var sessionID *string
	if p.SessionID != "" {
		sessionID = &p.SessionID
	}

	var parentRunID *string
	if p.ParentRunID != "" {
		parentRunID = &p.ParentRunID
	}

	const q = `
		INSERT INTO runs (id, task_id, task_desc, status, complexity, provider, started_at, session_id, project, parent_run_id, branch)
		VALUES (?, ?, ?, 'running', ?, ?, ?, ?, ?, ?, ?)`

	if _, err := tx.Exec(q,
		ev.RunID,
		p.TaskID,
		p.TaskDescription,
		p.Complexity,
		p.Provider,
		ev.Timestamp.Format(timestampFmt),
		sessionID,
		p.Project,
		parentRunID,
		p.Branch,
	); err != nil {
		return fmt.Errorf("dashboard/writer: insertar en runs (run.start): %w", err)
	}
	return nil
}

func handleRunEnd(tx *sql.Tx, ev instrumentation.Event) error {
	p, err := marshalPayload[instrumentation.RunEndPayload](ev)
	if err != nil {
		return err
	}

	const q = `
		UPDATE runs
		SET status        = ?,
		    ended_at      = ?,
		    duration_ms   = ?,
		    total_tokens  = ?,
		    files_touched = ?,
		    agents_count  = ?
		WHERE id = ?`

	res, err := tx.Exec(q,
		p.Status,
		ev.Timestamp.Format(timestampFmt),
		p.DurationMs,
		p.TotalTokens,
		p.TotalFiles,
		p.AgentsCompleted+p.AgentsFailed,
		ev.RunID,
	)
	if err != nil {
		return fmt.Errorf("dashboard/writer: actualizar runs (run.end): %w", err)
	}
	if err := expectRowsAffected(res, "runs", ev.RunID); err != nil {
		return err
	}

	return insertRunProjects(tx, ev.RunID)
}

// insertRunProjects queries DISTINCT file paths for the run and inserts a row
// per unique project into run_projects. No-op when no file paths are found.
func insertRunProjects(tx *sql.Tx, runID string) error {
	rows, err := tx.Query(
		`SELECT DISTINCT path FROM files WHERE run_id = ? AND path LIKE '/%'`,
		runID,
	)
	if err != nil {
		return fmt.Errorf("dashboard/writer: consultar files para run_projects: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	seen := make(map[string]struct{})
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return fmt.Errorf("dashboard/writer: escanear path para run_projects: %w", err)
		}
		project := extractProjectFromPath(path)
		if project == "" {
			continue
		}
		seen[project] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("dashboard/writer: iterar files para run_projects: %w", err)
	}
	rows.Close() //nolint:errcheck

	const insertProject = `INSERT OR IGNORE INTO run_projects (run_id, project) VALUES (?, ?)`
	for project := range seen {
		if _, err := tx.Exec(insertProject, runID, project); err != nil {
			return fmt.Errorf("dashboard/writer: insertar en run_projects: %w", err)
		}
	}
	return nil
}

func handleAgentStart(tx *sql.Tx, ev instrumentation.Event) error {
	p, err := marshalPayload[instrumentation.AgentStartPayload](ev)
	if err != nil {
		return err
	}

	dependsOnJSON, err := json.Marshal(p.DependsOn)
	if err != nil {
		return fmt.Errorf("dashboard/writer: serializar depends_on: %w", err)
	}

	const q = `
		INSERT INTO agents (run_id, agent_id, agent_role, sequence, depends_on, model, status, started_at)
		VALUES (?, ?, ?, ?, ?, ?, 'running', ?)`

	if _, err := tx.Exec(q,
		ev.RunID,
		p.AgentID,
		p.AgentRole,
		p.Sequence,
		string(dependsOnJSON),
		p.Model,
		ev.Timestamp.Format(timestampFmt),
	); err != nil {
		return fmt.Errorf("dashboard/writer: insertar en agents (agent.start): %w", err)
	}
	return nil
}

func handleAgentEnd(tx *sql.Tx, ev instrumentation.Event) error {
	p, err := marshalPayload[instrumentation.AgentEndPayload](ev)
	if err != nil {
		return err
	}

	const maxOutputBytes = 100 * 1024 // 100 KB

	output := p.Output
	if len(output) > maxOutputBytes {
		output = output[:maxOutputBytes]
		// Retroceder si cortamos a mitad de un carácter UTF-8 multibyte.
		for len(output) > 0 && output[len(output)-1]&0xC0 == 0x80 {
			output = output[:len(output)-1]
		}
		// Si el último byte es un líder UTF-8 incompleto, eliminarlo también.
		if len(output) > 0 && output[len(output)-1]&0xC0 == 0xC0 {
			output = output[:len(output)-1]
		}
	}

	const updateAgent = `
		UPDATE agents
		SET status        = ?,
		    ended_at      = ?,
		    duration_ms   = ?,
		    tokens_input  = ?,
		    tokens_output = ?,
		    tokens_total  = ?,
		    exit_code     = ?,
		    output        = ?
		WHERE run_id = ? AND agent_id = ?`

	res, err := tx.Exec(updateAgent,
		p.Status,
		ev.Timestamp.Format(timestampFmt),
		p.DurationMs,
		p.TokensInput,
		p.TokensOutput,
		p.TokensTotal,
		p.ExitCode,
		output,
		ev.RunID,
		p.AgentID,
	)
	if err != nil {
		return fmt.Errorf("dashboard/writer: actualizar agents (agent.end): %w", err)
	}
	if err := expectRowsAffected(res, "agents", p.AgentID); err != nil {
		return err
	}

	const insertFile = `INSERT INTO files (run_id, agent_id, path, operation) VALUES (?, ?, ?, ?)`
	for _, path := range p.FilesTouched {
		if _, err := tx.Exec(insertFile, ev.RunID, p.AgentID, path, "touched"); err != nil {
			return fmt.Errorf("dashboard/writer: insertar en files (agent.end): %w", err)
		}
	}
	return nil
}

func handleAgentError(tx *sql.Tx, ev instrumentation.Event) error {
	p, err := marshalPayload[instrumentation.AgentErrorPayload](ev)
	if err != nil {
		return err
	}

	const q = `
		UPDATE agents
		SET status      = 'failed',
		    ended_at    = ?,
		    duration_ms = ?,
		    error_msg   = ?
		WHERE run_id = ? AND agent_id = ?`

	res, err := tx.Exec(q,
		ev.Timestamp.Format(timestampFmt),
		p.DurationMs,
		p.Error,
		ev.RunID,
		p.AgentID,
	)
	if err != nil {
		return fmt.Errorf("dashboard/writer: actualizar agents (agent.error): %w", err)
	}
	return expectRowsAffected(res, "agents", p.AgentID)
}

func handleFileTouched(tx *sql.Tx, ev instrumentation.Event) error {
	p, err := marshalPayload[instrumentation.FileTouchedPayload](ev)
	if err != nil {
		return err
	}

	var diff *string
	if p.Diff != "" {
		diff = &p.Diff
	}

	const q = `INSERT INTO files (run_id, agent_id, path, operation, diff) VALUES (?, ?, ?, ?, ?)`
	if _, err := tx.Exec(q, ev.RunID, p.AgentID, p.Path, p.Operation, diff); err != nil {
		return fmt.Errorf("dashboard/writer: insertar en files (file.touched): %w", err)
	}
	return nil
}

func handleQAScore(tx *sql.Tx, ev instrumentation.Event) error {
	p, err := marshalPayload[instrumentation.QAScorePayload](ev)
	if err != nil {
		return err
	}

	const q = `UPDATE runs SET qa_score = ? WHERE id = ?`
	res, err := tx.Exec(q, p.Score, ev.RunID)
	if err != nil {
		return fmt.Errorf("dashboard/writer: actualizar runs (qa.score): %w", err)
	}
	return expectRowsAffected(res, "runs", ev.RunID)
}

func handleToolUse(tx *sql.Tx, ev instrumentation.Event) error {
	p, err := marshalPayload[instrumentation.ToolUsePayload](ev)
	if err != nil {
		return err
	}

	toolInput := string(p.ToolInput)
	source := p.Source
	if source == "" {
		source = "native"
	}
	var mcpServer *string
	if p.MCPServer != "" {
		mcpServer = &p.MCPServer
	}

	const q = `INSERT INTO tool_uses (run_id, agent_id, tool_name, tool_input, timestamp, source, mcp_server, input_size_bytes) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	if _, err := tx.Exec(q, ev.RunID, p.AgentID, p.ToolName, toolInput,
		ev.Timestamp.Format(timestampFmt), source, mcpServer, p.InputSizeBytes,
	); err != nil {
		return fmt.Errorf("dashboard/writer: insertar en tool_uses: %w", err)
	}
	return nil
}

func handleRunError(tx *sql.Tx, ev instrumentation.Event) error {
	p, err := marshalPayload[instrumentation.RunErrorPayload](ev)
	if err != nil {
		return err
	}

	const q = `UPDATE runs SET status = 'failed', error_reason = ?, ended_at = ? WHERE id = ?`
	_, err = tx.Exec(q, p.ErrorReason,
		ev.Timestamp.Format(timestampFmt),
		ev.RunID,
	)
	if err != nil {
		return fmt.Errorf("dashboard/writer: actualizar runs (run.error): %w", err)
	}
	return nil
}

func handleTaskCreated(tx *sql.Tx, ev instrumentation.Event) error {
	p, err := marshalPayload[instrumentation.TaskCreatedPayload](ev)
	if err != nil {
		return err
	}

	const q = `INSERT INTO tasks (run_id, task_id, title, status, created_at) VALUES (?, ?, ?, 'pending', ?)`
	if _, err := tx.Exec(q, ev.RunID, p.TaskID, p.Title,
		ev.Timestamp.Format(timestampFmt),
	); err != nil {
		return fmt.Errorf("dashboard/writer: insertar en tasks: %w", err)
	}
	return nil
}

func handleTaskCompleted(tx *sql.Tx, ev instrumentation.Event) error {
	p, err := marshalPayload[instrumentation.TaskCompletedPayload](ev)
	if err != nil {
		return err
	}

	const q = `UPDATE tasks SET status = 'completed', completed_at = ? WHERE run_id = ? AND task_id = ?`
	_, err = tx.Exec(q,
		ev.Timestamp.Format(timestampFmt),
		ev.RunID, p.TaskID,
	)
	if err != nil {
		return fmt.Errorf("dashboard/writer: actualizar tasks (task.completed): %w", err)
	}
	return nil
}

// UpdateToolUseDuration sets duration_ms on the most recent tool_uses row that
// matches run_id + tool_name + agent_id and has duration_ms IS NULL.
// Best-effort: if no matching row is found the call is a no-op (no error).
func (w *EventWriter) UpdateToolUseDuration(runID, toolName, agentID string, durationMs int64) error {
	const q = `
		UPDATE tool_uses SET duration_ms = ?
		WHERE rowid = (
			SELECT rowid FROM tool_uses
			WHERE run_id = ? AND tool_name = ? AND agent_id = ? AND duration_ms IS NULL
			ORDER BY timestamp DESC
			LIMIT 1
		)`
	_, err := w.db.Exec(q, durationMs, runID, toolName, agentID)
	if err != nil {
		return fmt.Errorf("dashboard/writer: actualizar duration_ms en tool_uses: %w", err)
	}
	return nil
}

func handleUserPrompt(tx *sql.Tx, ev instrumentation.Event) error {
	p, err := marshalPayload[instrumentation.UserPromptPayload](ev)
	if err != nil {
		return err
	}

	const q = `
		INSERT INTO prompts (run_id, sequence, prompt, timestamp)
		VALUES (?, COALESCE((SELECT MAX(sequence) FROM prompts WHERE run_id = ?), 0) + 1, ?, ?)`
	if _, err := tx.Exec(q,
		ev.RunID,
		ev.RunID,
		p.Prompt,
		ev.Timestamp.Format(timestampFmt),
	); err != nil {
		return fmt.Errorf("dashboard/writer: insertar en prompts: %w", err)
	}
	return nil
}
