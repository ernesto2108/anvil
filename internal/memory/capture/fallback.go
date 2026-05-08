package capture

import (
	"fmt"
	"strings"

	"github.com/ernesto2108/anvil/internal/memory"
	"github.com/ernesto2108/anvil/internal/memory/transcript"
)

// keywordsOnlyDigest builds a DigestDraft from the transcript metadata without
// calling any LLM. Used when --no-llm is set or when the Haiku call times out.
func keywordsOnlyDigest(td transcript.TranscriptDigest) memory.DigestDraft {
	summary := buildFallbackSummary(td)

	decisions := td.Decisions
	if decisions == nil {
		decisions = []string{}
	}

	return memory.DigestDraft{
		Summary:   summary,
		Decisions: decisions,
		EdgeCases: []string{},
		Errors:    []string{},
	}
}

// buildFallbackSummary constructs a short human-readable summary from the
// fields available in TranscriptDigest without LLM assistance.
func buildFallbackSummary(td transcript.TranscriptDigest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Direct session (%d turns)", td.TurnCount)
	if len(td.ToolsUsed) > 0 {
		fmt.Fprintf(&b, ". Tools used: %s", strings.Join(td.ToolsUsed, ", "))
	}
	if td.HasErrors {
		b.WriteString(". Had errors.")
	}
	if !td.StartedAt.IsZero() && !td.EndedAt.IsZero() {
		dur := td.EndedAt.Sub(td.StartedAt).Round(1e9) // round to seconds
		fmt.Fprintf(&b, ". Duration: %s", dur)
	}
	return b.String()
}
