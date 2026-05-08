package capture

import (
	"context"
	"database/sql"

	"github.com/ernesto2108/anvil/internal/memory"
	"github.com/ernesto2108/anvil/internal/memory/transcript"
)

// TranscriptSummarizer is the interface satisfied by haiku.Client for
// transcript-based summarization. Defined here to avoid an import cycle:
// capture → haiku would require haiku → capture.
type TranscriptSummarizer interface {
	SummarizeTranscript(ctx context.Context, td transcript.TranscriptDigest) (memory.DigestDraft, error)
}

// Orchestrator wires together transcript parsing, project resolution, and
// digest persistence for a single capture run.
type Orchestrator struct {
	Summarizer TranscriptSummarizer
	DB         *sql.DB
}

// NewOrchestrator constructs an Orchestrator. summarizer may be nil when
// operating in --no-llm mode; callers must handle that case before invoking
// Capture.
func NewOrchestrator(summarizer TranscriptSummarizer, db *sql.DB) *Orchestrator {
	return &Orchestrator{
		Summarizer: summarizer,
		DB:         db,
	}
}
