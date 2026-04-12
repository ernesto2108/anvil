package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ernesto2108/anvil/internal/orchestrator"
)

// ClaudeRunner executes agent nodes by spawning `claude --print` subprocesses.
type ClaudeRunner struct {
	WorkDir  string
	Model    string
	TaskDesc string // the high-level task description passed via --task flag
}

// New creates a ClaudeRunner ready to execute pipeline nodes.
func New(workDir, model, taskDesc string) *ClaudeRunner {
	return &ClaudeRunner{
		WorkDir:  workDir,
		Model:    model,
		TaskDesc: taskDesc,
	}
}

// RunAgent executes the given node by spawning `claude --print` and returns the result.
// upstream contains the completed results of dependency nodes, used to build
// context in the prompt. It is safe for concurrent use.
func (r *ClaudeRunner) RunAgent(ctx context.Context, node orchestrator.Node, upstream map[string]orchestrator.AgentResult) (orchestrator.AgentResult, error) {
	start := time.Now()

	prompt := r.buildPrompt(node, upstream)

	args := []string{"--print", "--bare", "-p", prompt}
	if r.Model != "" {
		args = append(args, "--model", r.Model)
	}

	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = r.WorkDir

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

// buildPrompt constructs the prompt for the agent based on its role and
// outputs from upstream dependencies.
func (r *ClaudeRunner) buildPrompt(node orchestrator.Node, upstream map[string]orchestrator.AgentResult) string {
	var b strings.Builder

	fmt.Fprintf(&b, "You are the %s agent.\n", node.Role)
	fmt.Fprintf(&b, "Task: %s\n\n", r.TaskDesc)

	// Inject outputs from completed dependencies.
	if len(upstream) > 0 {
		b.WriteString("## Context from previous agents\n\n")
		for _, depID := range node.DependsOn {
			if res, ok := upstream[depID]; ok && res.Output != "" {
				fmt.Fprintf(&b, "### Output from %s\n\n%s\n\n", depID, res.Output)
			}
		}
	}

	return b.String()
}
