package runner

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/ernesto2108/anvil/internal/orchestrator"
)

func TestNew(t *testing.T) {
	r := New("/tmp/work", "opus", "deploy the app")

	if r.WorkDir != "/tmp/work" {
		t.Errorf("WorkDir = %q, want %q", r.WorkDir, "/tmp/work")
	}
	if r.Model != "opus" {
		t.Errorf("Model = %q, want %q", r.Model, "opus")
	}
	if r.TaskDesc != "deploy the app" {
		t.Errorf("TaskDesc = %q, want %q", r.TaskDesc, "deploy the app")
	}
}

func TestBuildPrompt_NoUpstream(t *testing.T) {
	r := New("/tmp", "", "write tests")
	node := orchestrator.Node{ID: "dev", Role: "developer"}

	prompt := r.buildPrompt(node, nil)

	if !strings.Contains(prompt, "You are the developer agent.") {
		t.Errorf("prompt missing role line, got:\n%s", prompt)
	}
	if !strings.Contains(prompt, "Task: write tests") {
		t.Errorf("prompt missing task line, got:\n%s", prompt)
	}
	if strings.Contains(prompt, "Context from previous agents") {
		t.Error("prompt should not contain upstream context section when upstream is nil")
	}
}

func TestBuildPrompt_EmptyUpstream(t *testing.T) {
	r := New("/tmp", "", "write tests")
	node := orchestrator.Node{ID: "dev", Role: "developer", DependsOn: []string{"pm"}}

	prompt := r.buildPrompt(node, map[string]orchestrator.AgentResult{})

	if strings.Contains(prompt, "Context from previous agents") {
		t.Error("prompt should not contain upstream context when map is empty")
	}
}

func TestBuildPrompt_WithUpstream(t *testing.T) {
	r := New("/tmp", "", "build feature X")
	node := orchestrator.Node{
		ID:        "dev",
		Role:      "developer",
		DependsOn: []string{"pm", "arch"},
	}
	upstream := map[string]orchestrator.AgentResult{
		"pm":   {NodeID: "pm", Output: "PRD content here"},
		"arch": {NodeID: "arch", Output: "Architecture doc"},
	}

	prompt := r.buildPrompt(node, upstream)

	if !strings.Contains(prompt, "## Context from previous agents") {
		t.Error("prompt missing upstream context header")
	}
	if !strings.Contains(prompt, "### Output from pm") {
		t.Error("prompt missing pm output section")
	}
	if !strings.Contains(prompt, "PRD content here") {
		t.Error("prompt missing pm output content")
	}
	if !strings.Contains(prompt, "### Output from arch") {
		t.Error("prompt missing arch output section")
	}
	if !strings.Contains(prompt, "Architecture doc") {
		t.Error("prompt missing arch output content")
	}
}

func TestBuildPrompt_UpstreamSkipsEmptyOutput(t *testing.T) {
	r := New("/tmp", "", "task")
	node := orchestrator.Node{
		ID:        "dev",
		Role:      "developer",
		DependsOn: []string{"pm"},
	}
	upstream := map[string]orchestrator.AgentResult{
		"pm": {NodeID: "pm", Output: ""},
	}

	prompt := r.buildPrompt(node, upstream)

	if strings.Contains(prompt, "### Output from pm") {
		t.Error("prompt should skip upstream entries with empty output")
	}
}

func TestBuildPrompt_UpstreamOnlyIncludesDependencies(t *testing.T) {
	r := New("/tmp", "", "task")
	node := orchestrator.Node{
		ID:        "dev",
		Role:      "developer",
		DependsOn: []string{"pm"},
	}
	upstream := map[string]orchestrator.AgentResult{
		"pm":   {NodeID: "pm", Output: "pm output"},
		"arch": {NodeID: "arch", Output: "arch output not a dependency"},
	}

	prompt := r.buildPrompt(node, upstream)

	if !strings.Contains(prompt, "### Output from pm") {
		t.Error("prompt should include pm (a dependency)")
	}
	if strings.Contains(prompt, "### Output from arch") {
		t.Error("prompt should NOT include arch (not in DependsOn)")
	}
}

// ---------------------------------------------------------------------------
// RunAgent tests — these use a helper binary pattern to avoid spawning real
// `claude` processes. We override exec.CommandContext via a test helper.
// ---------------------------------------------------------------------------

// TestRunAgent_Success uses a real command (echo) to verify the success path.
func TestRunAgent_Success(t *testing.T) {
	if _, err := exec.LookPath("echo"); err != nil {
		t.Skip("echo not available")
	}

	r := &ClaudeRunner{
		WorkDir:  t.TempDir(),
		TaskDesc: "test task",
	}

	// We can't easily swap the "claude" binary, so we create a shell script
	// that acts as a fake claude.
	fakeClaude := createFakeClaude(t, "#!/bin/sh\necho 'agent output here'")
	t.Setenv("PATH", fakeClaude+":"+os.Getenv("PATH"))

	node := orchestrator.Node{ID: "dev", Role: "developer"}
	result, err := r.RunAgent(context.Background(), node, nil)

	if err != nil {
		t.Fatalf("RunAgent returned error: %v", err)
	}
	if result.Status != orchestrator.NodeSuccess {
		t.Errorf("Status = %q, want %q", result.Status, orchestrator.NodeSuccess)
	}
	if result.NodeID != "dev" {
		t.Errorf("NodeID = %q, want %q", result.NodeID, "dev")
	}
	if result.Output != "agent output here" {
		t.Errorf("Output = %q, want %q", result.Output, "agent output here")
	}
	if result.DurationMs <= 0 {
		t.Error("DurationMs should be positive")
	}
}

func TestRunAgent_Failure(t *testing.T) {
	fakeClaude := createFakeClaude(t, "#!/bin/sh\necho 'partial output' >&1\necho 'some error' >&2\nexit 1")
	t.Setenv("PATH", fakeClaude+":"+os.Getenv("PATH"))

	r := &ClaudeRunner{
		WorkDir:  t.TempDir(),
		TaskDesc: "test task",
	}

	node := orchestrator.Node{ID: "qa", Role: "qa"}
	result, err := r.RunAgent(context.Background(), node, nil)

	if err == nil {
		t.Fatal("RunAgent should return error on non-zero exit")
	}
	if result.Status != orchestrator.NodeFailed {
		t.Errorf("Status = %q, want %q", result.Status, orchestrator.NodeFailed)
	}
	if result.Output != "partial output" {
		t.Errorf("Output = %q, want %q", result.Output, "partial output")
	}
	if result.Error == nil {
		t.Error("result.Error should be non-nil")
	}
	if !strings.Contains(result.Error.Error(), "some error") {
		t.Errorf("result.Error should contain stderr, got: %v", result.Error)
	}
}

func TestRunAgent_ContextCancelled(t *testing.T) {
	fakeClaude := createFakeClaude(t, "#!/bin/sh\nsleep 30\necho 'should not reach'")
	t.Setenv("PATH", fakeClaude+":"+os.Getenv("PATH"))

	r := &ClaudeRunner{
		WorkDir:  t.TempDir(),
		TaskDesc: "test task",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	node := orchestrator.Node{ID: "dev", Role: "developer"}
	_, err := r.RunAgent(ctx, node, nil)

	if err == nil {
		t.Fatal("RunAgent should return error when context is cancelled")
	}
}

func TestRunAgent_WithModel(t *testing.T) {
	// Verify the model flag is passed to the command by checking args.
	// The fake claude script writes its args to a file so we can inspect them.
	tmpDir := t.TempDir()
	argsFile := tmpDir + "/args.txt"
	script := "#!/bin/sh\necho \"$@\" > " + argsFile + "\necho 'ok'"
	fakeClaude := createFakeClaude(t, script)
	t.Setenv("PATH", fakeClaude+":"+os.Getenv("PATH"))

	r := &ClaudeRunner{
		WorkDir:  t.TempDir(),
		Model:    "sonnet",
		TaskDesc: "test",
	}

	node := orchestrator.Node{ID: "dev", Role: "developer"}
	_, err := r.RunAgent(context.Background(), node, nil)
	if err != nil {
		t.Fatalf("RunAgent returned error: %v", err)
	}

	argsBytes, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read args file: %v", err)
	}
	args := string(argsBytes)

	if !strings.Contains(args, "--model sonnet") {
		t.Errorf("expected --model sonnet in args, got: %s", args)
	}
}

// createFakeClaude creates a temporary directory with a fake "claude" script
// and returns the directory path (to prepend to PATH).
func createFakeClaude(t *testing.T, script string) string {
	t.Helper()
	dir := t.TempDir()
	path := dir + "/claude"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to create fake claude: %v", err)
	}
	return dir
}
