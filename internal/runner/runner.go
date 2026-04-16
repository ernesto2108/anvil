package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ernesto2108/anvil/internal/orchestrator"
)

// TaskContext holds the enriched metadata gathered during preflight.
type TaskContext struct {
	Objective     string
	Stack         string
	Files         string
	Complexity    string
	MemoryContext string // pre-formatted digest summaries from previous runs
}

// ClaudeRunner executes agent nodes by spawning `claude --print` subprocesses.
//
// RunID is the parent Anvil run ID that owns the lifecycle of all spawned agents.
// It is propagated to each subprocess via the ANVIL_PARENT_RUN_ID env var so
// that Claude Code hooks (`anvil emit`) attach their telemetry to the same run
// instead of creating sibling orphan runs in the dashboard.
type ClaudeRunner struct {
	WorkDir string
	Model   string
	RunID   string
	Task    TaskContext
}

// New creates a ClaudeRunner ready to execute pipeline nodes.
func New(workDir, model, runID string, task TaskContext) *ClaudeRunner {
	return &ClaudeRunner{
		WorkDir: workDir,
		Model:   model,
		RunID:   runID,
		Task:    task,
	}
}

// RunAgent executes the given node by spawning `claude --print` and returns the result.
// upstream contains the completed results of dependency nodes, used to build
// context in the prompt. It is safe for concurrent use.
func (r *ClaudeRunner) RunAgent(ctx context.Context, node orchestrator.Node, upstream map[string]orchestrator.AgentResult) (orchestrator.AgentResult, error) {
	start := time.Now()

	prompt := r.buildPrompt(node, upstream)

	// --permission-mode acceptEdits lets the agent write/edit files without
	// the interactive permission prompt that would otherwise block a non-
	// interactive `--print` session.
	args := []string{"--print", "--agent", node.Role, "--permission-mode", "acceptEdits", "-p", prompt}
	if r.Model != "" {
		args = append(args, "--model", r.Model)
	}

	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = r.WorkDir

	// Propagate parent run identity to the subprocess. Claude Code hooks
	// (configured globally) call `anvil emit`, which inherits this env and
	// attaches events to the parent run instead of creating a new one.
	cmd.Env = append(os.Environ(),
		"ANVIL_PARENT_RUN_ID="+r.RunID,
		"ANVIL_AGENT_ID="+node.ID,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	durationMs := time.Since(start).Milliseconds()

	output := strings.TrimSpace(stdout.String())

	if err != nil {
		return orchestrator.AgentResult{
			NodeID:     node.ID,
			Status:     orchestrator.NodeFailed,
			DurationMs: durationMs,
			Output:     output,
			Error:      fmt.Errorf("runner: agent %q failed: %w\nstderr: %s", node.ID, err, stderr.String()),
		}, err
	}

	return orchestrator.AgentResult{
		NodeID:     node.ID,
		Status:     orchestrator.NodeSuccess,
		DurationMs: durationMs,
		Output:     output,
	}, nil
}

// buildPrompt constructs a structured prompt for the agent.
func (r *ClaudeRunner) buildPrompt(node orchestrator.Node, upstream map[string]orchestrator.AgentResult) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Complexity: %s\n", r.Task.Complexity)
	fmt.Fprintf(&b, "Stack: %s\n", r.Task.Stack)
	fmt.Fprintf(&b, "Mode: normal\n")
	fmt.Fprintf(&b, "Objective: %s\n", r.Task.Objective)
	fmt.Fprintf(&b, "\nFiles to change:\n%s\n", r.Task.Files)

	if r.Task.MemoryContext != "" {
		fmt.Fprintf(&b, "\n## Relevant memories from previous runs\n\n%s\n", r.Task.MemoryContext)
	}

	// Inject outputs from completed dependencies.
	if len(upstream) > 0 {
		b.WriteString("\n## Context from previous agents\n\n")
		for _, depID := range node.DependsOn {
			if res, ok := upstream[depID]; ok && res.Output != "" {
				fmt.Fprintf(&b, "### Output from %s\n\n%s\n\n", depID, res.Output)
			}
		}
	}

	return b.String()
}
