package transcript

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// --- JSONL event types ---

// transcriptEvent is a single line in a Claude Code JSONL transcript.
type transcriptEvent struct {
	Type      string          `json:"type"`
	UUID      string          `json:"uuid"`
	Cwd       string          `json:"cwd"`
	Timestamp string          `json:"timestamp"`
	Message   *transcriptMsg  `json:"message,omitempty"`
}

type transcriptMsg struct {
	Role    string           `json:"role"`
	Content json.RawMessage  `json:"content"`
}

// contentBlock is a single item in a message.content array.
type contentBlock struct {
	Type    string          `json:"type"`
	Name    string          `json:"name,omitempty"`    // tool_use
	IsError bool            `json:"is_error,omitempty"` // tool_result
	Text    string          `json:"text,omitempty"`    // text block
}

// ParseFile opens the JSONL transcript at path and delegates to Parse.
func ParseFile(path string) (TranscriptDigest, error) {
	f, err := os.Open(path)
	if err != nil {
		return TranscriptDigest{}, fmt.Errorf("transcript: open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	td, err := Parse(f)
	if err != nil {
		return TranscriptDigest{}, err
	}
	td.Path = path
	return td, nil
}

// Parse reads JSONL from r and builds a TranscriptDigest. Lines that are not
// valid JSON are skipped with a log message — a partial transcript is better
// than a hard error.
func Parse(r io.Reader) (TranscriptDigest, error) {
	var td TranscriptDigest

	toolsSeen := make(map[string]struct{})
	decisionLinesCandidates := []string{}

	scanner := bufio.NewScanner(r)
	// Some transcripts may have long lines (large diffs in content). Expand the
	// buffer to 10 MB to avoid "token too long" errors on those lines.
	const maxScanBuf = 10 * 1024 * 1024
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, maxScanBuf)

	var lastTimestamp time.Time

	for scanner.Scan() {
		raw := scanner.Bytes()
		if len(raw) == 0 {
			continue
		}

		var ev transcriptEvent
		if err := json.Unmarshal(raw, &ev); err != nil {
			log.Printf("transcript: skip malformed line: %v", err)
			continue
		}

		// Parse timestamp.
		var ts time.Time
		if ev.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339Nano, ev.Timestamp); err == nil {
				ts = t
			} else if t, err := time.Parse(time.RFC3339, ev.Timestamp); err == nil {
				ts = t
			}
		}

		// Track start/end times.
		if !ts.IsZero() {
			if td.StartedAt.IsZero() || ts.Before(td.StartedAt) {
				td.StartedAt = ts
			}
			if ts.After(lastTimestamp) {
				lastTimestamp = ts
			}
		}

		// Extract SessionID from the first system event or first uuid.
		if td.SessionID == "" && ev.UUID != "" {
			td.SessionID = ev.UUID
		}

		// Extract Cwd from the first event that has it.
		if td.Cwd == "" && ev.Cwd != "" {
			td.Cwd = ev.Cwd
		}

		// Process message events.
		if ev.Message == nil {
			continue
		}

		msg := ev.Message

		// Count human turns.
		if msg.Role == "user" {
			td.TurnCount++
		}

		// Parse content array.
		var blocks []contentBlock
		if len(msg.Content) > 0 {
			// Content can be a string or an array. Try array first.
			if err := json.Unmarshal(msg.Content, &blocks); err != nil {
				// Try string — wrap it as a single text block.
				var s string
				if err2 := json.Unmarshal(msg.Content, &s); err2 == nil {
					blocks = []contentBlock{{Type: "text", Text: s}}
				}
			}
		}

		for _, b := range blocks {
			switch b.Type {
			case "tool_use":
				if b.Name != "" {
					toolsSeen[b.Name] = struct{}{}
				}
			case "tool_result":
				if b.IsError {
					td.HasErrors = true
				}
			case "text":
				// Collect assistant text lines as candidates for decision extraction.
				if msg.Role == "assistant" && b.Text != "" {
					decisionLinesCandidates = append(decisionLinesCandidates, strings.Split(b.Text, "\n")...)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return TranscriptDigest{}, fmt.Errorf("transcript: scan: %w", err)
	}

	td.EndedAt = lastTimestamp

	// Deduplicate tools.
	if len(toolsSeen) > 0 {
		td.ToolsUsed = make([]string, 0, len(toolsSeen))
		for name := range toolsSeen {
			td.ToolsUsed = append(td.ToolsUsed, name)
		}
	}

	// Extract decisions from assistant text.
	td.Decisions = extractDecisions(decisionLinesCandidates)

	return td, nil
}
