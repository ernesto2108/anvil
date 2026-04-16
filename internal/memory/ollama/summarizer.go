package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ernesto2108/anvil/internal/memory"
)

// Summarizer implements memory.Summarizer using a local Ollama chat model.
type Summarizer struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewSummarizer creates an Ollama-based summarizer.
// Defaults: baseURL "http://localhost:11434", model "mistral".
func NewSummarizer(baseURL, model string) *Summarizer {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "mistral"
	}
	return &Summarizer{
		baseURL:    baseURL,
		model:      model,
		httpClient: &http.Client{},
	}
}

// Model returns the model identifier used for summarization.
func (s *Summarizer) Model() string {
	return s.model
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
	Options  chatOptions   `json:"options"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatOptions struct {
	Temperature float64 `json:"temperature"`
	NumPredict  int     `json:"num_predict"`
}

type chatResponse struct {
	Message chatMessage `json:"message"`
}

// Summarize sends agent outputs to the local Ollama model and returns a structured digest draft.
func (s *Summarizer) Summarize(ctx context.Context, agentOutputs []memory.AgentOutput) (memory.DigestDraft, error) {
	prompt := buildSummarizePrompt(agentOutputs)

	reqBody, err := json.Marshal(chatRequest{
		Model: s.model,
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
		Stream: false,
		Options: chatOptions{
			Temperature: 0,
			NumPredict:  1024,
		},
	})
	if err != nil {
		return memory.DigestDraft{}, fmt.Errorf("ollama summarizer: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/api/chat", bytes.NewReader(reqBody))
	if err != nil {
		return memory.DigestDraft{}, fmt.Errorf("ollama summarizer: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return memory.DigestDraft{}, fmt.Errorf("ollama summarizer: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return memory.DigestDraft{}, fmt.Errorf("ollama summarizer: unexpected status %d", resp.StatusCode)
	}

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return memory.DigestDraft{}, fmt.Errorf("ollama summarizer: decode response: %w", err)
	}

	if result.Message.Content == "" {
		return memory.DigestDraft{}, fmt.Errorf("ollama summarizer: empty response")
	}

	return parseDraftJSON(result.Message.Content)
}

func buildSummarizePrompt(outputs []memory.AgentOutput) string {
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

func parseDraftJSON(text string) (memory.DigestDraft, error) {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var d draftJSON
	if err := json.Unmarshal([]byte(text), &d); err != nil {
		return memory.DigestDraft{}, fmt.Errorf("ollama summarizer: parse draft JSON: %w", err)
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
