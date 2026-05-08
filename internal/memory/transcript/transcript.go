package transcript

import "time"

// TranscriptDigest holds the structured data extracted from a Claude Code
// JSONL transcript file. It is the input to capture.Capture and to
// haiku.Client.SummarizeTranscript.
type TranscriptDigest struct {
	// Path is the absolute path to the source JSONL file.
	Path string
	// SessionID is extracted from the first system event's uuid field.
	SessionID string
	// Cwd is the working directory recorded in the transcript.
	Cwd string
	// StartedAt is the timestamp of the first event.
	StartedAt time.Time
	// EndedAt is the timestamp of the last event.
	EndedAt time.Time
	// TurnCount is the number of human turns in the conversation.
	TurnCount int
	// Decisions are deduped assistant lines that contain decision keywords.
	Decisions []string
	// ToolsUsed are deduplicated tool names from tool_use content blocks.
	ToolsUsed []string
	// HasErrors is true when any tool_result with is_error or any error
	// indicator was found in the transcript.
	HasErrors bool
}
