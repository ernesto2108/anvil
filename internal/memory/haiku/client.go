package haiku

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ernesto2108/anvil/internal/memory"
)

const defaultModel = "claude-haiku-4-5"

// Client implements memory.Summarizer using the Anthropic Messages API.
type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewClient creates a Haiku summarizer client.
func NewClient(apiKey, model string) *Client {
	if model == "" {
		model = defaultModel
	}
	return &Client{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{},
	}
}

// Model returns the model identifier used for summarization.
func (c *Client) Model() string {
	return c.model
}

// --- Anthropic Messages API types ---

type messagesRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messagesResponse struct {
	Content []contentBlock `json:"content"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// --- Summarize ---

// Summarize sends agent outputs to Haiku and returns a structured digest draft.
func (c *Client) Summarize(ctx context.Context, agentOutputs []memory.AgentOutput) (memory.DigestDraft, error) {
	prompt := buildPrompt(agentOutputs)

	reqBody, err := json.Marshal(messagesRequest{
		Model:     c.model,
		MaxTokens: 1024,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return memory.DigestDraft{}, fmt.Errorf("haiku: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(reqBody))
	if err != nil {
		return memory.DigestDraft{}, fmt.Errorf("haiku: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return memory.DigestDraft{}, fmt.Errorf("haiku: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return memory.DigestDraft{}, fmt.Errorf("haiku: unexpected status %d", resp.StatusCode)
	}

	var result messagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return memory.DigestDraft{}, fmt.Errorf("haiku: decode response: %w", err)
	}

	if len(result.Content) == 0 {
		return memory.DigestDraft{}, fmt.Errorf("haiku: empty response")
	}

	return parseDraft(result.Content[0].Text)
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
	// Strip markdown fences if present despite instructions.
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var d draftJSON
	if err := json.Unmarshal([]byte(text), &d); err != nil {
		return memory.DigestDraft{}, fmt.Errorf("haiku: parse draft JSON: %w", err)
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
