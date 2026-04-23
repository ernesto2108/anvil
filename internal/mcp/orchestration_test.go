package mcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/ernesto2108/anvil/pkg/config"
)

// setupTestDB creates an in-memory SQLite database with the runs and agents tables.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`CREATE TABLE runs (
		id TEXT PRIMARY KEY,
		task_desc TEXT,
		status TEXT NOT NULL DEFAULT '',
		complexity TEXT DEFAULT '',
		provider TEXT DEFAULT '',
		project TEXT DEFAULT '',
		started_at TEXT DEFAULT '',
		ended_at TEXT DEFAULT '',
		duration_ms INTEGER DEFAULT 0,
		files_touched INTEGER DEFAULT 0,
		agents_count INTEGER DEFAULT 0,
		qa_score REAL,
		parent_run_id TEXT,
		task_source TEXT DEFAULT '',
		task_id TEXT DEFAULT '',
		pipeline TEXT NOT NULL DEFAULT '',
		stack TEXT NOT NULL DEFAULT '',
		pipeline_snapshot TEXT
	)`)
	if err != nil {
		t.Fatalf("create runs table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE agents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id TEXT NOT NULL,
		agent_id TEXT NOT NULL,
		agent_role TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT '',
		output TEXT,
		started_at TEXT DEFAULT '',
		ended_at TEXT DEFAULT '',
		duration_ms INTEGER DEFAULT 0,
		sequence INTEGER DEFAULT 0,
		files_touched_list TEXT NOT NULL DEFAULT '[]'
	)`)
	if err != nil {
		t.Fatalf("create agents table: %v", err)
	}

	return db
}

// newTestServer creates a Server with an in-memory DB and a temp dir as RepoDir.
func newTestServer(t *testing.T, db *sql.DB) (*Server, string) {
	t.Helper()
	tmpDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tmpDir, "pipelines"), 0o755)
	return &Server{
		cfg: &config.App{RepoDir: tmpDir},
		db:  db,
	}, tmpDir
}

// writePipelineYAML writes a pipeline YAML file in the given pipelines directory.
func writePipelineYAML(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, "pipelines", name+".yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write pipeline yaml: %v", err)
	}
}

var runIDPattern = regexp.MustCompile(`^r_\d{8}_\d{6}_[0-9a-f]{4}$`)

func TestStartOrchestration_Success(t *testing.T) {
	db := setupTestDB(t)
	srv, _ := newTestServer(t, db)

	args := map[string]any{
		"objective":  "implement feature X",
		"stack":      "Go",
		"complexity": "medium",
	}

	out, err := srv.startOrchestration(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}

	runID, ok := result["run_id"].(string)
	if !ok || runID == "" {
		t.Fatalf("missing or empty run_id in output")
	}
	if !runIDPattern.MatchString(runID) {
		t.Errorf("run_id %q does not match expected format r_YYYYMMDD_HHMMSS_xxxx", runID)
	}

	// Verify DB row.
	var taskSource, status string
	row := db.QueryRow(`SELECT task_source, status FROM runs WHERE id = ?`, runID)
	if err := row.Scan(&taskSource, &status); err != nil {
		t.Fatalf("query run row: %v", err)
	}
	if taskSource != "conversational" {
		t.Errorf("task_source = %q, want 'conversational'", taskSource)
	}
	if status != "running" {
		t.Errorf("status = %q, want 'running'", status)
	}
}

func TestStartOrchestration_WithPipeline(t *testing.T) {
	db := setupTestDB(t)
	srv, tmpDir := newTestServer(t, db)

	pipelineYAML := `nodes:
  - id: dev
    role: developer
    depends_on: []
    gate: false
  - id: tester
    role: tester
    depends_on: [dev]
    gate: false
`
	writePipelineYAML(t, tmpDir, "my-pipeline", pipelineYAML)

	args := map[string]any{
		"objective":  "build something",
		"stack":      "Go",
		"complexity": "small",
		"pipeline":   "my-pipeline",
	}

	out, err := srv.startOrchestration(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}

	runID := result["run_id"].(string)

	// Verify pipeline_snapshot in DB.
	var snapshotRaw sql.NullString
	if err := db.QueryRow(`SELECT pipeline_snapshot FROM runs WHERE id = ?`, runID).Scan(&snapshotRaw); err != nil {
		t.Fatalf("query snapshot: %v", err)
	}
	if !snapshotRaw.Valid || snapshotRaw.String == "" {
		t.Fatal("pipeline_snapshot is empty")
	}

	var nodes []pipelineNode
	if err := json.Unmarshal([]byte(snapshotRaw.String), &nodes); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].Role != "developer" {
		t.Errorf("nodes[0].Role = %q, want 'developer'", nodes[0].Role)
	}
	if nodes[1].Role != "tester" {
		t.Errorf("nodes[1].Role = %q, want 'tester'", nodes[1].Role)
	}
}

func TestStartOrchestration_PipelineNotFound(t *testing.T) {
	db := setupTestDB(t)
	srv, tmpDir := newTestServer(t, db)

	// Write an existing pipeline so the error message can list it.
	writePipelineYAML(t, tmpDir, "existing-pipeline", "nodes: []\n")

	args := map[string]any{
		"objective":  "do something",
		"stack":      "Go",
		"complexity": "small",
		"pipeline":   "nonexistent",
	}

	_, err := srv.startOrchestration(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for missing pipeline, got nil")
	}
	errMsg := err.Error()
	if !regexp.MustCompile(`nonexistent`).MatchString(errMsg) {
		t.Errorf("error %q should mention the missing pipeline name", errMsg)
	}
	if !regexp.MustCompile(`existing-pipeline`).MatchString(errMsg) {
		t.Errorf("error %q should list available pipelines", errMsg)
	}
}

func TestSaveStep_Success(t *testing.T) {
	db := setupTestDB(t)
	srv, _ := newTestServer(t, db)

	// Create a running run first.
	_, err := db.Exec(`INSERT INTO runs (id, task_desc, status, task_source, pipeline, stack, complexity, project, started_at)
		VALUES ('r_test_001', 'test run', 'running', 'conversational', '', 'Go', 'small', 'anvil', '2024-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("insert run: %v", err)
	}

	longOutput := "This is the full agent output without any truncation. " +
		"It can be arbitrarily long and the handler should store it entirely."

	args := map[string]any{
		"run_id":  "r_test_001",
		"role":    "developer",
		"status":  "success",
		"output":  longOutput,
	}

	out, err := srv.saveStep(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}

	if _, ok := result["step_id"]; !ok {
		t.Error("missing step_id in response")
	}

	// Verify DB: sequence=1 for first step.
	var sequence int
	var storedOutput string
	if err := db.QueryRow(`SELECT sequence, output FROM agents WHERE run_id = 'r_test_001'`).Scan(&sequence, &storedOutput); err != nil {
		t.Fatalf("query agent: %v", err)
	}
	if sequence != 1 {
		t.Errorf("sequence = %d, want 1", sequence)
	}
	if storedOutput != longOutput {
		t.Errorf("output truncated: got len=%d, want len=%d", len(storedOutput), len(longOutput))
	}
}

func TestSaveStep_RunNotRunning(t *testing.T) {
	db := setupTestDB(t)
	srv, _ := newTestServer(t, db)

	_, err := db.Exec(`INSERT INTO runs (id, task_desc, status, task_source, pipeline, stack, complexity, project, started_at)
		VALUES ('r_done_001', 'done run', 'done', 'conversational', '', 'Go', 'small', 'anvil', '2024-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("insert run: %v", err)
	}

	args := map[string]any{
		"run_id":  "r_done_001",
		"role":    "tester",
		"status":  "success",
		"output":  "some output",
	}

	out, err := srv.saveStep(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var errResult map[string]any
	if err := json.Unmarshal([]byte(out), &errResult); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if _, ok := errResult["error"]; !ok {
		t.Errorf("expected error key in response, got: %s", out)
	}
}

func TestSaveStep_RunNotFound(t *testing.T) {
	db := setupTestDB(t)
	srv, _ := newTestServer(t, db)

	args := map[string]any{
		"run_id":  "r_nonexistent",
		"role":    "developer",
		"status":  "success",
		"output":  "some output",
	}

	out, err := srv.saveStep(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var errResult map[string]any
	if err := json.Unmarshal([]byte(out), &errResult); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if _, ok := errResult["error"]; !ok {
		t.Errorf("expected error key in response, got: %s", out)
	}
}

func TestSaveStep_IncrementsCounts(t *testing.T) {
	db := setupTestDB(t)
	srv, _ := newTestServer(t, db)

	_, err := db.Exec(`INSERT INTO runs (id, task_desc, status, task_source, pipeline, stack, complexity, project, started_at)
		VALUES ('r_counts_001', 'count test', 'running', 'conversational', '', 'Go', 'small', 'anvil', '2024-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("insert run: %v", err)
	}

	// Save first step with 2 files.
	args1 := map[string]any{
		"run_id":        "r_counts_001",
		"role":          "developer",
		"status":        "success",
		"output":        "step 1 output",
		"files_touched": []any{"file_a.go", "file_b.go"},
	}
	if _, err := srv.saveStep(context.Background(), args1); err != nil {
		t.Fatalf("saveStep 1: %v", err)
	}

	// Save second step with 1 file.
	args2 := map[string]any{
		"run_id":        "r_counts_001",
		"role":          "tester",
		"status":        "success",
		"output":        "step 2 output",
		"files_touched": []any{"file_c.go"},
	}
	if _, err := srv.saveStep(context.Background(), args2); err != nil {
		t.Fatalf("saveStep 2: %v", err)
	}

	var agentsCount, filesTouched int
	if err := db.QueryRow(`SELECT agents_count, files_touched FROM runs WHERE id = 'r_counts_001'`).Scan(&agentsCount, &filesTouched); err != nil {
		t.Fatalf("query run counts: %v", err)
	}
	if agentsCount != 2 {
		t.Errorf("agents_count = %d, want 2", agentsCount)
	}
	if filesTouched != 3 {
		t.Errorf("files_touched = %d, want 3", filesTouched)
	}
}

func TestLoadOrchestration_FullState(t *testing.T) {
	db := setupTestDB(t)
	srv, _ := newTestServer(t, db)

	snapshot := `[{"id":"dev","role":"developer","depends_on":[],"gate":false},{"id":"tester","role":"tester","depends_on":["dev"],"gate":false}]`

	_, err := db.Exec(`INSERT INTO runs (id, task_desc, status, task_source, pipeline, stack, complexity, project, started_at, pipeline_snapshot)
		VALUES ('r_full_001', 'full state test', 'running', 'conversational', 'my-pipeline', 'Go', 'small', 'anvil', '2024-01-01T00:00:00Z', ?)`,
		snapshot)
	if err != nil {
		t.Fatalf("insert run: %v", err)
	}

	// Insert a completed developer step.
	_, err = db.Exec(`INSERT INTO agents (run_id, agent_id, agent_role, status, output, sequence, files_touched_list, duration_ms, started_at, ended_at)
		VALUES ('r_full_001', 'a_developer_1', 'developer', 'success', 'dev output', 1, '[]', 100, '2024-01-01T00:00:00Z', '2024-01-01T00:01:00Z')`)
	if err != nil {
		t.Fatalf("insert agent: %v", err)
	}

	args := map[string]any{
		"run_id":  "r_full_001",
		"project": "anvil",
	}

	out, err := srv.loadOrchestration(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}

	if result["run_id"] != "r_full_001" {
		t.Errorf("run_id = %v, want 'r_full_001'", result["run_id"])
	}

	steps, ok := result["steps"].([]any)
	if !ok || len(steps) != 1 {
		t.Errorf("expected 1 step, got %v", result["steps"])
	}

	// developer completed, so tester should be pending.
	pendingRoles, ok := result["pending_roles"].([]any)
	if !ok {
		t.Fatalf("pending_roles not a list: %v", result["pending_roles"])
	}
	if len(pendingRoles) != 1 || pendingRoles[0] != "tester" {
		t.Errorf("pending_roles = %v, want ['tester']", pendingRoles)
	}
}

func TestLoadOrchestration_Last(t *testing.T) {
	db := setupTestDB(t)
	srv, _ := newTestServer(t, db)

	// Insert a pipeline run (should NOT be resolved as "last" for conversational).
	_, err := db.Exec(`INSERT INTO runs (id, task_desc, status, task_source, pipeline, stack, complexity, project, started_at)
		VALUES ('r_pipeline_001', 'pipeline run', 'done', 'pipeline', 'some-pipeline', 'Go', 'small', 'anvil', '2024-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("insert pipeline run: %v", err)
	}

	// Insert a conversational run (newer).
	_, err = db.Exec(`INSERT INTO runs (id, task_desc, status, task_source, pipeline, stack, complexity, project, started_at)
		VALUES ('r_conv_001', 'conversational run', 'running', 'conversational', '', 'Go', 'small', 'anvil', '2024-01-02T00:00:00Z')`)
	if err != nil {
		t.Fatalf("insert conversational run: %v", err)
	}

	args := map[string]any{
		"run_id":  "last",
		"project": "anvil",
	}

	out, err := srv.loadOrchestration(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}

	if result["run_id"] != "r_conv_001" {
		t.Errorf("run_id = %v, want 'r_conv_001' (should resolve to conversational run, not pipeline run)", result["run_id"])
	}
}

func TestLoadOrchestration_NotFound(t *testing.T) {
	db := setupTestDB(t)
	srv, _ := newTestServer(t, db)

	args := map[string]any{
		"run_id":  "last",
		"project": "anvil",
	}

	out, err := srv.loadOrchestration(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}

	if _, ok := result["error"]; !ok {
		t.Errorf("expected error key in response, got: %s", out)
	}
}
