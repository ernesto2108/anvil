package mcp

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// pipelineNode is a single node parsed from a pipeline YAML file.
type pipelineNode struct {
	ID        string   `yaml:"id"        json:"id"`
	Role      string   `yaml:"role"       json:"role"`
	DependsOn []string `yaml:"depends_on" json:"depends_on"`
	Gate      bool     `yaml:"gate"       json:"gate"`
}

// pipelineFile is the top-level structure of a pipeline YAML file.
type pipelineFile struct {
	Nodes []pipelineNode `yaml:"nodes"`
}

// startOrchestration creates a new conversational orchestration run and
// persists it to the runs table with task_source='conversational'.
func (s *Server) startOrchestration(_ context.Context, args map[string]any) (string, error) {
	if s.db == nil {
		return "", fmt.Errorf("database not available")
	}

	objective := strArg(args, "objective", "")
	if objective == "" {
		return "", fmt.Errorf("objective is required")
	}
	stack := strArg(args, "stack", "")
	if stack == "" {
		return "", fmt.Errorf("stack is required")
	}
	complexity := strArg(args, "complexity", "")
	if complexity == "" {
		return "", fmt.Errorf("complexity is required")
	}
	pipeline := strArg(args, "pipeline", "")

	// Generate run ID: r_YYYYMMDD_HHMMSS_xxxx
	now := time.Now().UTC()
	b := make([]byte, 2)
	_, _ = rand.Read(b)
	runID := fmt.Sprintf("r_%s_%04x", now.Format("20060102_150405"), b)

	// Resolve project from cwd.
	workDir, _ := os.Getwd()
	project := filepath.Base(workDir)

	// Handle optional pipeline snapshot.
	var pipelineName string
	var snapshotJSON sql.NullString

	if pipeline != "" {
		pipelineName = pipeline
		pipelinePath := filepath.Join(s.cfg.RepoDir, "pipelines", pipeline+".yaml")
		data, err := os.ReadFile(pipelinePath)
		if os.IsNotExist(err) {
			files, _ := filepath.Glob(filepath.Join(s.cfg.RepoDir, "pipelines", "*.yaml"))
			names := make([]string, 0, len(files))
			for _, f := range files {
				names = append(names, strings.TrimSuffix(filepath.Base(f), ".yaml"))
			}
			return "", fmt.Errorf("pipeline %q not found — available: %s", pipeline, strings.Join(names, ", "))
		}
		if err != nil {
			return "", fmt.Errorf("read pipeline %q: %w", pipeline, err)
		}

		var pf pipelineFile
		if err := yaml.Unmarshal(data, &pf); err != nil {
			return "", fmt.Errorf("parse pipeline %q: %w", pipeline, err)
		}

		// Normalize nil DependsOn to empty slice for consistent JSON output.
		for i := range pf.Nodes {
			if pf.Nodes[i].DependsOn == nil {
				pf.Nodes[i].DependsOn = []string{}
			}
		}

		snapBytes, err := json.Marshal(pf.Nodes)
		if err != nil {
			return "", fmt.Errorf("serialize pipeline snapshot: %w", err)
		}
		snapshotJSON = sql.NullString{String: string(snapBytes), Valid: true}
	}

	createdAt := now.Format(time.RFC3339)

	_, err := s.db.ExecContext(context.Background(),
		`INSERT INTO runs (id, task_id, task_desc, status, task_source, pipeline, stack, complexity, pipeline_snapshot, project, started_at)
		 VALUES (?, '', ?, 'running', 'conversational', ?, ?, ?, ?, ?, ?)`,
		runID, objective, pipelineName, stack, complexity, snapshotJSON, project, createdAt,
	)
	if err != nil {
		return "", fmt.Errorf("insert run: %w", err)
	}

	result := map[string]any{
		"run_id":     runID,
		"created_at": createdAt,
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

// saveStep persists an agent step within an active conversational run.
// It requires the run to exist and have status='running'.
func (s *Server) saveStep(_ context.Context, args map[string]any) (string, error) {
	if s.db == nil {
		return "", fmt.Errorf("database not available")
	}

	runID := strArg(args, "run_id", "")
	if runID == "" {
		return "", fmt.Errorf("run_id is required")
	}
	role := strArg(args, "role", "")
	if role == "" {
		return "", fmt.Errorf("role is required")
	}
	status := strArg(args, "status", "")
	if status == "" {
		return "", fmt.Errorf("status is required")
	}
	output := strArg(args, "output", "")
	if output == "" {
		return "", fmt.Errorf("output is required")
	}

	durationMs := intArg(args, "duration_ms", 0)

	// Parse files_touched (optional array of strings).
	var filesTouched []string
	if v, ok := args["files_touched"]; ok {
		if arr, ok := v.([]any); ok {
			for _, item := range arr {
				if s, ok := item.(string); ok {
					filesTouched = append(filesTouched, s)
				}
			}
		}
	}
	if filesTouched == nil {
		filesTouched = []string{}
	}

	// Verify run exists and is running.
	var runStatus string
	err := s.db.QueryRowContext(context.Background(),
		`SELECT status FROM runs WHERE id = ?`, runID).Scan(&runStatus)
	if err == sql.ErrNoRows {
		return fmt.Sprintf(`{"error": "run %s not found"}`, runID), nil
	}
	if err != nil {
		return "", fmt.Errorf("query run: %w", err)
	}
	if runStatus != "running" {
		return fmt.Sprintf(`{"error": "run %s is not running (status: %s)"}`, runID, runStatus), nil
	}

	// Calculate next sequence number.
	var maxSeq sql.NullInt64
	_ = s.db.QueryRowContext(context.Background(),
		`SELECT MAX(sequence) FROM agents WHERE run_id = ?`, runID).Scan(&maxSeq)
	sequence := int(maxSeq.Int64) + 1

	agentID := fmt.Sprintf("a_%s_%d", role, sequence)
	now := time.Now().UTC().Format(time.RFC3339)

	filesTouchedJSON, _ := json.Marshal(filesTouched)

	_, err = s.db.ExecContext(context.Background(),
		`INSERT INTO agents (run_id, agent_id, agent_role, status, output, sequence, files_touched_list, duration_ms, started_at, ended_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		runID, agentID, role, status, output, sequence,
		string(filesTouchedJSON), durationMs, now, now,
	)
	if err != nil {
		return "", fmt.Errorf("insert agent: %w", err)
	}

	_, err = s.db.ExecContext(context.Background(),
		`UPDATE runs SET agents_count = agents_count + 1, files_touched = files_touched + ? WHERE id = ?`,
		len(filesTouched), runID,
	)
	if err != nil {
		return "", fmt.Errorf("update run counts: %w", err)
	}

	result := map[string]any{
		"step_id":    agentID,
		"node_index": sequence,
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

// loadOrchestration loads a conversational run with all its steps and
// computes pending_roles from the pipeline snapshot.
// Use run_id="last" to load the most recent conversational run.
func (s *Server) loadOrchestration(_ context.Context, args map[string]any) (string, error) {
	if s.db == nil {
		return "", fmt.Errorf("database not available")
	}

	runID := strArg(args, "run_id", "")
	if runID == "" {
		return "", fmt.Errorf("run_id is required")
	}

	workDir, _ := os.Getwd()
	project := strArg(args, "project", filepath.Base(workDir))

	// Resolve "last" scoped to conversational runs for this project.
	if runID == "last" {
		err := s.db.QueryRowContext(context.Background(),
			`SELECT id FROM runs WHERE task_source='conversational' AND project=? ORDER BY started_at DESC LIMIT 1`,
			project).Scan(&runID)
		if err == sql.ErrNoRows {
			return `{"error": "no conversational runs found"}`, nil
		}
		if err != nil {
			return "", fmt.Errorf("resolve last run: %w", err)
		}
	}

	// Load run row.
	var (
		taskDesc, status        string
		pipelineName, stack     sql.NullString
		complexity, snapshotRaw sql.NullString
		startedAt               string
		endedAt                 sql.NullString
	)
	err := s.db.QueryRowContext(context.Background(),
		`SELECT task_desc, status, pipeline, stack, complexity, pipeline_snapshot, started_at, ended_at
		 FROM runs WHERE id = ?`, runID).
		Scan(&taskDesc, &status, &pipelineName, &stack, &complexity, &snapshotRaw, &startedAt, &endedAt)
	if err == sql.ErrNoRows {
		return fmt.Sprintf(`{"error": "run %s not found"}`, runID), nil
	}
	if err != nil {
		return "", fmt.Errorf("query run: %w", err)
	}

	// Load agents ordered by sequence.
	rows, err := s.db.QueryContext(context.Background(),
		`SELECT agent_id, agent_role, status, output, files_touched_list, duration_ms, sequence
		 FROM agents WHERE run_id = ? ORDER BY sequence`, runID)
	if err != nil {
		return "", fmt.Errorf("query agents: %w", err)
	}
	defer rows.Close()

	type stepResult struct {
		StepID       string   `json:"step_id"`
		Role         string   `json:"role"`
		Status       string   `json:"status"`
		Output       string   `json:"output"`
		FilesTouched []string `json:"files_touched"`
		DurationMs   int64    `json:"duration_ms"`
		NodeIndex    int      `json:"node_index"`
	}

	var steps []stepResult
	completedRoles := map[string]bool{}

	for rows.Next() {
		var (
			agentID, role, agentStatus string
			output, ftList             sql.NullString
			durMs                      sql.NullInt64
			seq                        int
		)
		if err := rows.Scan(&agentID, &role, &agentStatus, &output, &ftList, &durMs, &seq); err != nil {
			continue
		}

		var files []string
		if ftList.Valid && ftList.String != "" {
			_ = json.Unmarshal([]byte(ftList.String), &files)
		}
		if files == nil {
			files = []string{}
		}

		steps = append(steps, stepResult{
			StepID:       agentID,
			Role:         role,
			Status:       agentStatus,
			Output:       output.String,
			FilesTouched: files,
			DurationMs:   durMs.Int64,
			NodeIndex:    seq,
		})

		if agentStatus == "success" {
			completedRoles[role] = true
		}
	}

	if steps == nil {
		steps = []stepResult{}
	}

	// Compute pending_roles from pipeline snapshot.
	pendingRoles := []string{}
	if snapshotRaw.Valid && snapshotRaw.String != "" {
		var nodes []pipelineNode
		if err := json.Unmarshal([]byte(snapshotRaw.String), &nodes); err == nil {
			for _, n := range nodes {
				// Skip gate nodes — they are not agent roles.
				if n.Role == "gate" {
					continue
				}
				if !completedRoles[n.Role] {
					pendingRoles = append(pendingRoles, n.Role)
				}
			}
		}
	}

	result := map[string]any{
		"run_id":        runID,
		"objective":     taskDesc,
		"pipeline":      pipelineName.String,
		"stack":         stack.String,
		"complexity":    complexity.String,
		"status":        status,
		"started_at":    startedAt,
		"ended_at":      endedAt.String,
		"steps":         steps,
		"pending_roles": pendingRoles,
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}
