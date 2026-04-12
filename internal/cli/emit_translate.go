package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/store"
	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// hookEnvelope is the top-level JSON structure that Claude Code sends to hooks
// via stdin. All fields are top-level (not nested).
type hookEnvelope struct {
	// Common fields (all events)
	SessionID     string `json:"session_id"`
	HookEventName string `json:"hook_event_name"`
	CWD           string `json:"cwd,omitempty"`

	// UserPromptSubmit
	Prompt string `json:"prompt,omitempty"`

	// SubagentStart / SubagentStop
	AgentID   string `json:"agent_id,omitempty"`
	AgentType string `json:"agent_type,omitempty"`

	// SubagentStop
	LastAssistantMessage string `json:"last_assistant_message,omitempty"`

	// PostToolUse
	ToolName  string          `json:"tool_name,omitempty"`
	ToolInput json.RawMessage `json:"tool_input,omitempty"`
}

// toolInputFile captures file_path and change content from Write/Edit tool_input.
type toolInputFile struct {
	FilePath  string `json:"file_path"`
	Content   string `json:"content,omitempty"`
	OldString string `json:"old_string,omitempty"`
	NewString string `json:"new_string,omitempty"`
}

const maxOutputBytes = 100 * 1024 // 100 KB

// translateHook parses the hook JSON and writes the corresponding Anvil events
// to the store. It resolves the run by session_id, creating an orphan run if
// needed (except for SessionStart which always creates a new run).
func translateHook(raw []byte, s *store.SQLiteStore) error {
	var env hookEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("parse hook JSON: %w", err)
	}

	if env.SessionID == "" {
		return fmt.Errorf("hook JSON missing session_id")
	}

	switch env.HookEventName {
	case "SessionStart":
		return handleSessionStart(s, env.SessionID, env.CWD)
	case "UserPromptSubmit":
		return handleUserPromptSubmit(s, env.SessionID, env.Prompt)
	case "SubagentStart":
		return handleSubagentStart(s, env.SessionID, env.AgentID, env.AgentType)
	case "SubagentStop":
		return handleSubagentStop(s, env.SessionID, env.AgentID, env.LastAssistantMessage)
	case "PostToolUse":
		return handlePostToolUse(s, env.SessionID, env.ToolName, env.ToolInput)
	case "Stop", "SessionEnd":
		return handleSessionEnd(s, env.SessionID)
	default:
		// Unknown hook type — ignore silently.
		return nil
	}
}

func handleSessionStart(s *store.SQLiteStore, sessionID, cwd string) error {
	runID, err := instrumentation.NewRunID()
	if err != nil {
		return err
	}

	payload := instrumentation.RunStartPayload{
		TaskID:    runID,
		Provider:  "claude-code",
		SessionID: sessionID,
		Project:   projectFromCWD(cwd),
	}

	ev, err := instrumentation.NewEvent(runID, instrumentation.EventRunStart, payload)
	if err != nil {
		return err
	}
	return s.WriteEvent(ev)
}

// projectFromCWD extracts the project name from a working directory path.
// "/Users/ernesto/projects/anvil" → "anvil"
func projectFromCWD(cwd string) string {
	if cwd == "" {
		return ""
	}
	// Remove trailing slash
	for len(cwd) > 1 && cwd[len(cwd)-1] == '/' {
		cwd = cwd[:len(cwd)-1]
	}
	// Extract last segment
	for i := len(cwd) - 1; i >= 0; i-- {
		if cwd[i] == '/' {
			return cwd[i+1:]
		}
	}
	return cwd
}

func handleUserPromptSubmit(s *store.SQLiteStore, sessionID, prompt string) error {
	runID, err := resolveOrCreateRun(s, sessionID)
	if err != nil {
		return err
	}
	if prompt == "" {
		return nil
	}
	return s.UpdateTaskDesc(runID, prompt)
}

func handleSubagentStart(s *store.SQLiteStore, sessionID, agentID, agentType string) error {
	runID, err := resolveOrCreateRun(s, sessionID)
	if err != nil {
		return err
	}

	if agentID == "" {
		agentID = fmt.Sprintf("agent-%s", sessionID[:8])
	}

	payload := instrumentation.AgentStartPayload{
		AgentID:   agentID,
		AgentRole: agentType,
	}

	ev, err := instrumentation.NewEvent(runID, instrumentation.EventAgentStart, payload)
	if err != nil {
		return err
	}
	return s.WriteEvent(ev)
}

func handleSubagentStop(s *store.SQLiteStore, sessionID, agentID, lastMessage string) error {
	runID, err := resolveOrCreateRun(s, sessionID)
	if err != nil {
		return err
	}

	if agentID == "" {
		agentID = fmt.Sprintf("agent-%s", sessionID[:8])
	}

	output := lastMessage
	if len(output) > maxOutputBytes {
		output = output[:maxOutputBytes]
	}

	// Calculate duration from agent start time.
	var durationMs int64
	if startedAt, ok := s.AgentStartedAt(runID, agentID); ok {
		if t, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", startedAt); err == nil {
			durationMs = time.Since(t).Milliseconds()
		}
	}

	payload := instrumentation.AgentEndPayload{
		AgentID:    agentID,
		Status:     "success",
		DurationMs: durationMs,
		Output:     output,
	}

	ev, err := instrumentation.NewEvent(runID, instrumentation.EventAgentEnd, payload)
	if err != nil {
		return err
	}
	return s.WriteEvent(ev)
}

func handlePostToolUse(s *store.SQLiteStore, sessionID, toolName string, toolInput json.RawMessage) error {
	// Only track Write and Edit tools.
	if toolName != "Write" && toolName != "Edit" {
		return nil
	}

	runID, err := resolveOrCreateRun(s, sessionID)
	if err != nil {
		return err
	}

	var ti toolInputFile
	if len(toolInput) > 0 {
		if err := json.Unmarshal(toolInput, &ti); err != nil {
			return fmt.Errorf("parse PostToolUse tool_input: %w", err)
		}
	}

	if ti.FilePath == "" {
		return nil
	}

	operation := "create"
	if toolName == "Edit" {
		operation = "modify"
	}

	// Associate file with the currently running agent (if any).
	agentID := s.ActiveAgentID(runID)

	// Build diff from tool_input content.
	var diff string
	const maxDiffBytes = 50 * 1024 // 50 KB
	if toolName == "Edit" && (ti.OldString != "" || ti.NewString != "") {
		diff = fmt.Sprintf("--- old\n+++ new\n- %s\n+ %s", ti.OldString, ti.NewString)
	} else if toolName == "Write" && ti.Content != "" {
		diff = ti.Content
	}
	if len(diff) > maxDiffBytes {
		diff = diff[:maxDiffBytes] + "\n... (truncated)"
	}

	payload := instrumentation.FileTouchedPayload{
		AgentID:   agentID,
		Path:      ti.FilePath,
		Operation: operation,
		Diff:      diff,
	}

	ev, err := instrumentation.NewEvent(runID, instrumentation.EventFileTouched, payload)
	if err != nil {
		return err
	}
	return s.WriteEvent(ev)
}

func handleSessionEnd(s *store.SQLiteStore, sessionID string) error {
	runID, err := s.ResolveRunBySession(sessionID)
	if err != nil {
		return err
	}
	if runID == "" {
		return nil
	}

	payload := instrumentation.RunEndPayload{
		Status: "success",
	}

	ev, err := instrumentation.NewEvent(runID, instrumentation.EventRunEnd, payload)
	if err != nil {
		return err
	}
	if err := s.WriteEvent(ev); err != nil {
		return err
	}

	// Overwrite totals AFTER run.end (which sets them to 0).
	return s.ComputeRunTotals(runID)
}

// resolveOrCreateRun finds the run for a session_id, or creates an orphan run
// if none exists (orphan recovery).
func resolveOrCreateRun(s *store.SQLiteStore, sessionID string) (string, error) {
	runID, err := s.ResolveRunBySession(sessionID)
	if err != nil {
		return "", err
	}
	if runID != "" {
		return runID, nil
	}

	runID, err = instrumentation.NewRunID()
	if err != nil {
		return "", err
	}

	payload := instrumentation.RunStartPayload{
		TaskID:          runID,
		TaskDescription: "(recovered session)",
		Provider:        "claude-code",
		SessionID:       sessionID,
	}

	ev, err := instrumentation.NewEvent(runID, instrumentation.EventRunStart, payload)
	if err != nil {
		return "", err
	}
	if err := s.WriteEvent(ev); err != nil {
		return "", err
	}
	return runID, nil
}
