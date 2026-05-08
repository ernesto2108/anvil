package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ernesto2108/anvil/internal/memory"
	"github.com/ernesto2108/anvil/internal/memory/transcript"
)

// Client implements memory.Summarizer using the local claude CLI.
// No API key required — inherits the authenticated Claude Code session.
type Client struct{}

// New creates a Claude CLI summarizer client.
func New() *Client {
	return &Client{}
}

// Model returns the model identifier used for summarization.
func (c *Client) Model() string {
	return "claude-cli"
}

// Summarize sends agent outputs to the claude CLI and returns a structured digest draft.
func (c *Client) Summarize(ctx context.Context, agentOutputs []memory.AgentOutput) (memory.DigestDraft, error) {
	prompt := buildPrompt(agentOutputs)
	text, err := c.run(ctx, prompt)
	if err != nil {
		return memory.DigestDraft{}, err
	}
	return parseDraft(text)
}

// SummarizeTranscript sends a compact transcript digest to the claude CLI and returns
// a structured DigestDraft.
func (c *Client) SummarizeTranscript(ctx context.Context, td transcript.TranscriptDigest) (memory.DigestDraft, error) {
	prompt := buildTranscriptPrompt(td)
	text, err := c.run(ctx, prompt)
	if err != nil {
		return memory.DigestDraft{}, err
	}
	return parseDraft(text)
}

func (c *Client) run(ctx context.Context, prompt string) (string, error) {
	cmd := exec.CommandContext(ctx, "claude", "-p", prompt)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if ok := isExitError(err, &exitErr); ok {
			return "", fmt.Errorf("claude: cli exited with code %d: %s", exitErr.ExitCode(), string(exitErr.Stderr))
		}
		return "", fmt.Errorf("claude: exec failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func isExitError(err error, target **exec.ExitError) bool {
	e, ok := err.(*exec.ExitError)
	if ok {
		*target = e
	}
	return ok
}

func buildTranscriptPrompt(td transcript.TranscriptDigest) string {
	var b strings.Builder
	b.WriteString("You are a technical summarizer for a developer tool.\n\n")
	b.WriteString("Given metadata from a Claude Code direct session, produce a structured summary.\n\n")
	fmt.Fprintf(&b, "## Session Metadata\n\n")
	fmt.Fprintf(&b, "- Session ID: %s\n", td.SessionID)
	fmt.Fprintf(&b, "- Turns: %d\n", td.TurnCount)
	fmt.Fprintf(&b, "- Had errors: %v\n", td.HasErrors)
	if len(td.ToolsUsed) > 0 {
		fmt.Fprintf(&b, "- Tools used: %s\n", strings.Join(td.ToolsUsed, ", "))
	}
	if len(td.Decisions) > 0 {
		b.WriteString("\n## Decision Signals\n\n")
		for _, d := range td.Decisions {
			fmt.Fprintf(&b, "- %s\n", d)
		}
	}
	b.WriteString(`
## Instructions

Produce a JSON object with exactly these fields:

{
  "summary": "2-3 sentence overview of the session based on the metadata",
  "decisions": ["each significant decision inferred from the decision signals"],
  "edge_cases": [],
  "errors": []
}

Rules:
- summary: factual, based only on provided metadata.
- decisions: 1-5 items inferred from the decision signals above.
- edge_cases and errors: use empty arrays unless obvious from metadata.
- Respond ONLY with the JSON object, no markdown fences, no explanation.`)

	return b.String()
}

func buildPrompt(outputs []memory.AgentOutput) string {
	var b strings.Builder
	b.WriteString(`You are a technical summarizer for an AI agent orchestration system.

Given the outputs from multiple agents that ran as part of a pipeline, produce a structured summary.

## Agent Outputs

`)
	for _, o := range outputs {
		fmt.Fprintf(&b, "### Agent: %s (%s)\n%s\n\n", o.AgentID, o.Role, o.Output)
	}

	b.WriteString(`## Instructions

Analyze all agent outputs and produce a JSON object with exactly these fields:

{
  "summary": "2-4 sentence overview of what the pipeline accomplished, including key files modified and functionality added/changed",
  "decisions": ["each significant technical decision made during the run, e.g. 'Used composition over inheritance for X'"],
  "edge_cases": ["each edge case identified or handled, e.g. 'Empty input returns nil instead of error'"],
  "errors": ["each error encountered and how it was resolved, or unresolved issues"]
}

Rules:
- summary: factual, no opinions. Mention specific files, functions, patterns.
- decisions: 3-7 items. Only non-obvious decisions, not "created a file".
- edge_cases: 0-5 items. Only if agents explicitly mentioned edge cases.
- errors: 0-5 items. Only actual errors/failures, not warnings.
- If a field has no items, use an empty array [].
- Respond ONLY with the JSON object, no markdown fences, no explanation.`)

	return b.String()
}

type draftJSON struct {
	Summary   string   `json:"summary"`
	Decisions []string `json:"decisions"`
	EdgeCases []string `json:"edge_cases"`
	Errors    []string `json:"errors"`
}

func parseDraft(text string) (memory.DigestDraft, error) {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var d draftJSON
	if err := json.Unmarshal([]byte(text), &d); err != nil {
		return memory.DigestDraft{}, fmt.Errorf("claude: parse draft JSON: %w", err)
	}

	if d.Decisions == nil {
		d.Decisions = []string{}
	}
	if d.EdgeCases == nil {
		d.EdgeCases = []string{}
	}
	if d.Errors == nil {
		d.Errors = []string{}
	}

	return memory.DigestDraft{
		Summary:   d.Summary,
		Decisions: d.Decisions,
		EdgeCases: d.EdgeCases,
		Errors:    d.Errors,
	}, nil
}
