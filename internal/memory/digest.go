package memory

import (
	"context"
	"time"
)

// Digest is the domain entity for a run summary.
type Digest struct {
	ID        string
	RunID     string
	Project   string
	Summary   string
	Decisions []string
	EdgeCases []string
	Errors    []string
	Embedding []float32 // nil when not generated
	ModelUsed string
	Source    string // 'auto' (digestCheckpointer) | 'manual' (dashboard) | 'handoff' (parsed from .handoff/<TASK-ID>.md)
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DigestDraft is the summarizer output before persisting.
type DigestDraft struct {
	Summary   string
	Decisions []string
	EdgeCases []string
	Errors    []string
}

// AgentOutput is input for the summarizer.
type AgentOutput struct {
	AgentID string
	Role    string
	Output  string
}

// SearchResult pairs a digest with its similarity score.
type SearchResult struct {
	Digest Digest
	Score  float64
}

// Embedder generates embedding vectors for text.
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// Summarizer generates a structured digest from agent outputs.
type Summarizer interface {
	Summarize(ctx context.Context, agentOutputs []AgentOutput) (DigestDraft, error)
}
