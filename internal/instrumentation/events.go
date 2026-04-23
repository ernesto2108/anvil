package instrumentation

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"
)

const SchemaVersion = "1"

// Event type constants.
const (
	EventRunStart    = "run.start"
	EventRunEnd      = "run.end"
	EventAgentStart  = "agent.start"
	EventAgentEnd    = "agent.end"
	EventAgentError  = "agent.error"
	EventFileTouched = "file.touched"
	EventQAScore     = "qa.score"

	EventOrchestratorStart = "orchestrator.start"
	EventOrchestratorGate  = "orchestrator.gate"

	// New hook-driven events
	EventToolUse          = "tool.use"
	EventRunError         = "run.error"
	EventTaskCreated      = "task.created"
	EventTaskCompleted    = "task.completed"
	EventCompaction       = "context.compacted"
	EventPermissionDenied = "permission.denied"
	EventUserPrompt       = "user.prompt"
)

// Event is the envelope for all instrumentation events.
type Event struct {
	SchemaVersion string          `json:"schema_version"`
	EventID       string          `json:"event_id"`
	RunID         string          `json:"run_id"`
	EventType     string          `json:"event_type"`
	Timestamp     time.Time       `json:"timestamp"`
	Payload       json.RawMessage `json:"payload"`
}

// EventWriter decouples the emitter from the persistence layer.
type EventWriter interface {
	WriteEvent(ev Event) error
}

// Emitter is the interface for emitting instrumentation events.
type Emitter interface {
	Emit(ev Event)
	Start()
	Stop()
}

// Payload types

type RunStartPayload struct {
	TaskID          string   `json:"task_id"`
	TaskDescription string   `json:"task_desc"`
	Complexity      string   `json:"complexity"`
	AgentsPlanned   []string `json:"agents_planned"`
	Provider        string   `json:"provider"`
	TriggeredBy     string   `json:"triggered_by"`
	SessionID       string   `json:"session_id,omitempty"`
	Project         string   `json:"project,omitempty"`
	ParentRunID     string   `json:"parent_run_id,omitempty"`
	Branch          string   `json:"branch,omitempty"`
}

type RunEndPayload struct {
	Status          string `json:"status"`
	DurationMs      int64  `json:"duration_ms"`
	TotalTokens     int    `json:"total_tokens"`
	TotalFiles      int    `json:"total_files_touched"`
	AgentsCompleted int    `json:"agents_completed"`
	AgentsFailed    int    `json:"agents_failed"`
}

type AgentStartPayload struct {
	AgentID   string   `json:"agent_id"`
	AgentRole string   `json:"agent_role"`
	Sequence  int      `json:"sequence"`
	DependsOn []string `json:"depends_on"`
	Model     string   `json:"model"`
}

type AgentEndPayload struct {
	AgentID      string   `json:"agent_id"`
	Status       string   `json:"status"`
	DurationMs   int64    `json:"duration_ms"`
	TokensInput  int      `json:"tokens_input"`
	TokensOutput int      `json:"tokens_output"`
	TokensTotal  int      `json:"tokens_total"`
	FilesTouched []string `json:"files_touched"`
	ExitCode     int      `json:"exit_code"`
	Output       string   `json:"output"`
}

type AgentErrorPayload struct {
	AgentID    string `json:"agent_id"`
	Error      string `json:"error"`
	StderrTail string `json:"stderr_tail"`
	DurationMs int64  `json:"duration_ms"`
}

type FileTouchedPayload struct {
	AgentID   string `json:"agent_id"`
	Path      string `json:"path"`
	Operation string `json:"operation"`
	Diff      string `json:"diff,omitempty"`
}

type QAScorePayload struct {
	AgentID  string             `json:"agent_id"`
	Score    float64            `json:"score"`
	MaxScore float64            `json:"max_score"`
	Criteria map[string]float64 `json:"criteria"`
	Notes    string             `json:"notes,omitempty"`
}

type OrchestratorStartPayload struct {
	NodeCount      int      `json:"node_count"`
	MaxConcurrency int      `json:"max_concurrency"`
	AgentIDs       []string `json:"agent_ids"`
}

type OrchestratorGatePayload struct {
	AgentID        string `json:"agent_id"`
	Decision       string `json:"decision"`
	WaitDurationMs int64  `json:"wait_duration_ms"`
}

// Hook-driven payload types

type ToolUsePayload struct {
	AgentID        string          `json:"agent_id"`
	ToolName       string          `json:"tool_name"`
	ToolInput      json.RawMessage `json:"tool_input,omitempty"`
	Source         string          `json:"source,omitempty"`         // "native" | "mcp"
	MCPServer      string          `json:"mcp_server,omitempty"`     // e.g. "pencil", "gh-mcp"
	InputSizeBytes int             `json:"input_size_bytes,omitempty"`
}

type RunErrorPayload struct {
	ErrorReason string `json:"error_reason"`
}

type TaskCreatedPayload struct {
	TaskID string `json:"task_id"`
	Title  string `json:"title"`
}

type TaskCompletedPayload struct {
	TaskID string `json:"task_id"`
}

type CompactionPayload struct {
	Type string `json:"type"` // "auto" | "manual"
}

type PermissionDeniedPayload struct {
	ToolName string `json:"tool_name"`
}

type UserPromptPayload struct {
	Prompt string `json:"prompt"`
}

// NewEvent creates a new Event with a UUID v4 event ID, UTC timestamp, and
// marshaled payload. Returns an error if payload marshaling fails.
func NewEvent(runID, eventType string, payload any) (Event, error) {
	id, err := newUUIDv4()
	if err != nil {
		return Event{}, fmt.Errorf("instrumentation: generate event id: %w", err)
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return Event{}, fmt.Errorf("instrumentation: marshal payload: %w", err)
	}

	return Event{
		SchemaVersion: SchemaVersion,
		EventID:       id,
		RunID:         runID,
		EventType:     eventType,
		Timestamp:     time.Now().UTC(),
		Payload:       raw,
	}, nil
}

// NewRunID generates a run ID in the format r_YYYYMMDD_HHmmss_XXXX.
func NewRunID() (string, error) {
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("instrumentation: generate run id: %w", err)
	}
	now := time.Now().UTC()
	return fmt.Sprintf("r_%s_%04x", now.Format("20060102_150405"), b), nil
}

// newUUIDv4 generates a random UUID v4 string using crypto/rand.
func newUUIDv4() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// Set version 4 and variant bits.
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%12x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:],
	), nil
}
