package runner

import (
	"bytes"
	"context"
	"encoding/json"
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
	WorkDir    string
	Model      string
	RunID      string
	Task       TaskContext
	MCPByRole  map[string][]byte // pre-built {"mcpServers":{...}} JSON per agent role
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

	// If MCP servers are declared for this agent role, write a temp config file
	// and pass --mcp-config so the subprocess only sees the allowed servers.
	mcpTmp, err := r.writeMCPConfig(node.Role)
	if err != nil {
		return orchestrator.AgentResult{}, fmt.Errorf("runner: write mcp config: %w", err)
	}
	if mcpTmp != "" {
		defer os.Remove(mcpTmp)
		args = append(args, "--mcp-config", mcpTmp)
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

	err = cmd.Run()
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

// writeMCPConfig writes the pre-built MCP JSON for the given role to a temp
// file and returns its path. Returns ("", nil) if no MCP config exists for
// the role — the caller must not pass --mcp-config in that case.
// The caller is responsible for removing the file after the subprocess exits.
func (r *ClaudeRunner) writeMCPConfig(role string) (string, error) {
	if len(r.MCPByRole) == 0 {
		return "", nil
	}
	data, ok := r.MCPByRole[role]
	if !ok || len(data) == 0 {
		return "", nil
	}
	f, err := os.CreateTemp("", "anvil-mcp-*.json")
	if err != nil {
		return "", fmt.Errorf("create temp mcp config: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return "", fmt.Errorf("write temp mcp config: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

// BuildMCPByRole constructs the MCPByRole map from a set of server definitions
// per agent role. defs maps agent role → server name → server entry.
// Each entry is serialized to the JSON structure Claude Code expects:
//
//	{"mcpServers": {"name": {"command": "...", "args": [...], "env": {...}}}}
func BuildMCPByRole(defs map[string]map[string]MCPEntry) (map[string][]byte, error) {
	if len(defs) == 0 {
		return nil, nil
	}
	result := make(map[string][]byte, len(defs))
	for role, servers := range defs {
		if len(servers) == 0 {
			continue
		}
		payload := map[string]any{"mcpServers": servers}
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal mcp config for role %q: %w", role, err)
		}
		result[role] = b
	}
	return result, nil
}

// MCPEntry is a single MCP server entry as expected by Claude Code's --mcp-config JSON.
type MCPEntry struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}
