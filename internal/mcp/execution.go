package mcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// runPipeline launches an Anvil pipeline preset in the background.
// It validates the pipeline exists, then exec's `anvil run` as a detached
// subprocess. Returns immediately without waiting for completion.
func (s *Server) runPipeline(_ context.Context, args map[string]any) (string, error) {
	pipeline := strArg(args, "pipeline", "")
	if pipeline == "" {
		return "", fmt.Errorf("pipeline is required")
	}
	task := strArg(args, "task", "")
	if task == "" {
		return "", fmt.Errorf("task is required")
	}
	stack := strArg(args, "stack", "")
	maxRetries := intArg(args, "max_retries", 2)
	maxCost := floatArg(args, "max_cost", 0.50)

	// Validate pipeline exists — only allow names from pipelines/ directory.
	pipelineFile := filepath.Join(s.cfg.RepoDir, "pipelines", pipeline+".yaml")
	if _, err := os.Stat(pipelineFile); os.IsNotExist(err) {
		// List available pipelines for a helpful error.
		files, _ := filepath.Glob(filepath.Join(s.cfg.RepoDir, "pipelines", "*.yaml"))
		names := make([]string, 0, len(files))
		for _, f := range files {
			names = append(names, strings.TrimSuffix(filepath.Base(f), ".yaml"))
		}
		return "", fmt.Errorf("pipeline %q not found — available: %s", pipeline, strings.Join(names, ", "))
	}

	// Build the command. -y auto-approves gates so the pipeline runs unattended.
	cmdArgs := []string{"run", pipeline, "--task", task, "-y",
		"--max-retries", strconv.Itoa(maxRetries),
		"--max-cost", strconv.FormatFloat(maxCost, 'f', 4, 64),
	}
	if stack != "" {
		cmdArgs = append(cmdArgs, "--stack", stack)
	}

	anvil, err := os.Executable()
	if err != nil {
		anvil = "anvil"
	}

	cmd := exec.Command(anvil, cmdArgs...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("launch pipeline: %w", err)
	}
	// Detach — we don't Wait(). The subprocess lives independently.
	go func() { _ = cmd.Wait() }()

	result := map[string]any{
		"status":      "launched",
		"pipeline":    pipeline,
		"task":        task,
		"pid":         cmd.Process.Pid,
		"max_retries": maxRetries,
		"max_cost":    maxCost,
		"note":        "Pipeline running in background. Use list_runs to check status.",
	}
	b, _ := json.MarshalIndent(result, "", "  ")
	return string(b), nil
}

// getRunStatus returns the status of a specific run, including its agents.
// Use run_id = "last" to fetch the most recent run.
func (s *Server) getRunStatus(_ context.Context, args map[string]any) (string, error) {
	runID := strArg(args, "run_id", "")
	if runID == "" {
		return "", fmt.Errorf("run_id is required")
	}

	db := s.db
	if db == nil {
		return "", fmt.Errorf("database not available")
	}

	// Resolve "last".
	if runID == "last" {
		var id string
		err := db.QueryRowContext(context.Background(),
			`SELECT id FROM runs ORDER BY started_at DESC LIMIT 1`).Scan(&id)
		if err == sql.ErrNoRows {
			return `{"error": "no runs found"}`, nil
		}
		if err != nil {
			return "", fmt.Errorf("resolve last run: %w", err)
		}
		runID = id
	}

	type runRow struct {
		ID           string  `json:"id"`
		TaskDesc     string  `json:"task"`
		Status       string  `json:"status"`
		Complexity   string  `json:"complexity"`
		Provider     string  `json:"provider"`
		Project      string  `json:"project"`
		StartedAt    string  `json:"started_at"`
		EndedAt      string  `json:"ended_at,omitempty"`
		DurationMs   int64   `json:"duration_ms"`
		FilesTouched int     `json:"files_touched"`
		AgentsCount  int     `json:"agents_count"`
		QAScore      float64 `json:"qa_score,omitempty"`
		Degraded     bool    `json:"degraded,omitempty"`
	}

	var r runRow
	var endedAt, complexity, provider, project sql.NullString
	var durationMs sql.NullInt64
	var filesTouched, agentsCount sql.NullInt64
	var qaScore sql.NullFloat64
	var degraded sql.NullInt64

	err := db.QueryRowContext(context.Background(),
		`SELECT id, task_desc, status, complexity, provider, project,
		        started_at, ended_at, duration_ms, files_touched, agents_count, qa_score, degraded
		 FROM runs WHERE id = ?`, runID).
		Scan(&r.ID, &r.TaskDesc, &r.Status, &complexity, &provider, &project,
			&r.StartedAt, &endedAt, &durationMs, &filesTouched, &agentsCount, &qaScore, &degraded)
	if err == sql.ErrNoRows {
		return fmt.Sprintf(`{"error": "run %q not found"}`, runID), nil
	}
	if err != nil {
		return "", fmt.Errorf("query run: %w", err)
	}
	r.Complexity = complexity.String
	r.Provider = provider.String
	r.Project = project.String
	r.EndedAt = endedAt.String
	r.DurationMs = durationMs.Int64
	r.FilesTouched = int(filesTouched.Int64)
	r.AgentsCount = int(agentsCount.Int64)
	r.QAScore = qaScore.Float64
	r.Degraded = degraded.Int64 == 1

	// Load agents for this run.
	type agentRow struct {
		AgentID   string `json:"agent_id"`
		Role      string `json:"role"`
		Status    string `json:"status"`
		StartedAt string `json:"started_at,omitempty"`
		EndedAt   string `json:"ended_at,omitempty"`
		DurationMs int64 `json:"duration_ms,omitempty"`
	}

	rows, err := db.QueryContext(context.Background(),
		`SELECT agent_id, agent_role, status, started_at, ended_at, duration_ms
		 FROM agents WHERE run_id = ? ORDER BY sequence`, runID)
	if err != nil {
		return "", fmt.Errorf("query agents: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var agents []agentRow
	for rows.Next() {
		var a agentRow
		var startedAt, endedAt sql.NullString
		var durMs sql.NullInt64
		if err := rows.Scan(&a.AgentID, &a.Role, &a.Status, &startedAt, &endedAt, &durMs); err != nil {
			continue
		}
		a.StartedAt = startedAt.String
		a.EndedAt = endedAt.String
		a.DurationMs = durMs.Int64
		agents = append(agents, a)
	}

	result := map[string]any{"run": r, "agents": agents}
	b, _ := json.MarshalIndent(result, "", "  ")
	return string(b), nil
}

// listRuns returns recent runs, optionally filtered by project or status.
func (s *Server) listRuns(_ context.Context, args map[string]any) (string, error) {
	db := s.db
	if db == nil {
		return "", fmt.Errorf("database not available")
	}

	workDir, _ := os.Getwd()
	project := strArg(args, "project", filepath.Base(workDir))
	status := strArg(args, "status", "")
	limit := intArg(args, "limit", 10)
	if limit > 50 {
		limit = 50
	}

	query := `SELECT id, task_desc, status, project, started_at, ended_at, duration_ms, files_touched, agents_count
	          FROM runs WHERE project = ?`
	qArgs := []any{project}
	if status != "" {
		query += " AND status = ?"
		qArgs = append(qArgs, status)
	}
	query += " ORDER BY started_at DESC LIMIT ?"
	qArgs = append(qArgs, limit)

	rows, err := db.QueryContext(context.Background(), query, qArgs...)
	if err != nil {
		return "", fmt.Errorf("query runs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	type runSummary struct {
		ID           string `json:"id"`
		Task         string `json:"task"`
		Status       string `json:"status"`
		Project      string `json:"project"`
		StartedAt    string `json:"started_at"`
		EndedAt      string `json:"ended_at,omitempty"`
		DurationMs   int64  `json:"duration_ms,omitempty"`
		FilesTouched int    `json:"files_touched,omitempty"`
		AgentsCount  int    `json:"agents_count,omitempty"`
	}

	var runs []runSummary
	for rows.Next() {
		var r runSummary
		var task, endedAt sql.NullString
		var durMs, files, agents sql.NullInt64
		if err := rows.Scan(&r.ID, &task, &r.Status, &r.Project,
			&r.StartedAt, &endedAt, &durMs, &files, &agents); err != nil {
			continue
		}
		r.Task = task.String
		r.EndedAt = endedAt.String
		r.DurationMs = durMs.Int64
		r.FilesTouched = int(files.Int64)
		r.AgentsCount = int(agents.Int64)
		runs = append(runs, r)
	}

	if runs == nil {
		runs = []runSummary{}
	}

	b, _ := json.MarshalIndent(runs, "", "  ")
	return string(b), nil
}

// getAgentOutput returns the output of a specific agent within a run.
func (s *Server) getAgentOutput(_ context.Context, args map[string]any) (string, error) {
	db := s.db
	if db == nil {
		return "", fmt.Errorf("database not available")
	}

	runID := strArg(args, "run_id", "")
	if runID == "" {
		return "", fmt.Errorf("run_id is required")
	}
	agentRole := strArg(args, "agent_role", "")
	if agentRole == "" {
		return "", fmt.Errorf("agent_role is required")
	}

	// Resolve "last".
	if runID == "last" {
		var id string
		err := db.QueryRowContext(context.Background(),
			`SELECT id FROM runs ORDER BY started_at DESC LIMIT 1`).Scan(&id)
		if err == sql.ErrNoRows {
			return `{"error": "no runs found"}`, nil
		}
		if err != nil {
			return "", fmt.Errorf("resolve last run: %w", err)
		}
		runID = id
	}

	const maxOutput = 50_000 // 50KB

	var agentID, status string
	var output sql.NullString
	var startedAt, endedAt sql.NullString
	var durMs sql.NullInt64

	err := db.QueryRowContext(context.Background(),
		`SELECT agent_id, status, output, started_at, ended_at, duration_ms
		 FROM agents WHERE run_id = ? AND agent_role = ?
		 ORDER BY sequence DESC LIMIT 1`,
		runID, agentRole).
		Scan(&agentID, &status, &output, &startedAt, &endedAt, &durMs)
	if err == sql.ErrNoRows {
		return fmt.Sprintf(`{"error": "agent %q not found in run %q"}`, agentRole, runID), nil
	}
	if err != nil {
		return "", fmt.Errorf("query agent: %w", err)
	}

	out := output.String
	truncated := false
	if len(out) > maxOutput {
		out = out[:maxOutput]
		truncated = true
	}

	result := map[string]any{
		"run_id":      runID,
		"agent_id":    agentID,
		"agent_role":  agentRole,
		"status":      status,
		"started_at":  startedAt.String,
		"ended_at":    endedAt.String,
		"duration_ms": durMs.Int64,
		"output":      out,
		"truncated":   truncated,
		"retrieved_at": time.Now().UTC().Format(time.RFC3339),
	}
	b, _ := json.MarshalIndent(result, "", "  ")
	return string(b), nil
}
