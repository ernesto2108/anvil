package cli

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
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

	// PostToolUse / PreToolUse / PermissionDenied
	ToolName  string          `json:"tool_name,omitempty"`
	ToolInput json.RawMessage `json:"tool_input,omitempty"`

	// StopFailure
	StopFailureReason string `json:"stop_failure_reason,omitempty"`

	// TaskCreated / TaskCompleted
	TaskID    string `json:"task_id,omitempty"`
	TaskTitle string `json:"task_title,omitempty"`

	// PostCompact
	CompactType string `json:"compact_type,omitempty"`
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
		// No-op: defer run creation to the first UserPromptSubmit so that
		// sessions opened but never used don't pollute the runs list.
		return nil
	case "UserPromptSubmit":
		return handleUserPromptSubmit(s, env.SessionID, env.Prompt, env.CWD)
	case "SubagentStart":
		return handleSubagentStart(s, env.SessionID, env.AgentID, env.AgentType, env.CWD)
	case "SubagentStop":
		return handleSubagentStop(s, env.SessionID, env.AgentID, env.LastAssistantMessage)
	case "PreToolUse":
		return handlePreToolUse(s, env.SessionID, env.ToolName, env.ToolInput, env.CWD)
	case "PostToolUse":
		return handlePostToolUse(s, env.SessionID, env.ToolName, env.ToolInput, env.CWD)
	case "StopFailure":
		return handleStopFailure(s, env.SessionID, env.StopFailureReason)
	case "TaskCreated":
		return handleTaskCreated(s, env.SessionID, env.TaskID, env.TaskTitle, env.CWD)
	case "TaskCompleted":
		return handleTaskCompleted(s, env.SessionID, env.TaskID)
	case "PostCompact":
		return handlePostCompact(s, env.SessionID, env.CompactType)
	case "PermissionDenied":
		return handlePermissionDenied(s, env.SessionID, env.ToolName)
	case "Stop":
		// Stop fires between turns (assistant finishes, waits for user).
		// It does NOT mean the session ended — ignore it.
		return nil
	case "SessionEnd":
		return handleSessionEnd(s, env.SessionID)
	default:
		// Unknown hook type — ignore silently.
		return nil
	}
}

// createRunForSession creates a new run linked to the given session.
// Called lazily on first real activity (prompt or agent), not on SessionStart.
func createRunForSession(s *store.SQLiteStore, sessionID, cwd string) (string, error) {
	runID, err := instrumentation.NewRunID()
	if err != nil {
		return "", err
	}

	payload := instrumentation.RunStartPayload{
		TaskID:    runID,
		Provider:  "claude-code",
		SessionID: sessionID,
		Project:   projectFromCWD(cwd),
		Branch:    branchFromCWD(cwd),
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

// branchFromCWD returns the current git branch for the given working directory.
// Returns "" if cwd is empty or git fails (not a repo, git not installed, etc.).
func branchFromCWD(cwd string) string {
	if cwd == "" {
		return ""
	}
	out, err := exec.Command("git", "-C", cwd, "branch", "--show-current").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func handleUserPromptSubmit(s *store.SQLiteStore, sessionID, prompt, cwd string) error {
	runID, err := resolveOrCreateRun(s, sessionID, cwd)
	if err != nil {
		return err
	}
	if prompt == "" {
		return nil
	}

	ev, err := instrumentation.NewEvent(runID, instrumentation.EventUserPrompt, instrumentation.UserPromptPayload{
		Prompt: prompt,
	})
	if err != nil {
		return err
	}
	if err := s.WriteEvent(ev); err != nil {
		return err
	}

	// UpdateTaskDesc is idempotent — only sets task_desc if currently empty,
	// so it naturally captures only the first prompt without needing to check sequence.
	return s.UpdateTaskDesc(runID, prompt)
}

func handleSubagentStart(s *store.SQLiteStore, sessionID, agentID, agentType, cwd string) error {
	runID, err := resolveOrCreateRun(s, sessionID, cwd)
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
	// SubagentStop should never fire without a prior SubagentStart that
	// already created the run, so CWD is unnecessary here.
	runID, err := resolveOrCreateRun(s, sessionID, "")
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

func handlePostToolUse(s *store.SQLiteStore, sessionID, toolName string, toolInput json.RawMessage, cwd string) error {
	// Only track Write and Edit tools.
	if toolName != "Write" && toolName != "Edit" {
		return nil
	}

	runID, err := resolveOrCreateRun(s, sessionID, cwd)
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

// --- New hook handlers -------------------------------------------------------

func handlePreToolUse(s *store.SQLiteStore, sessionID, toolName string, toolInput json.RawMessage, cwd string) error {
	// Skip Write/Edit — already captured as file.touched via PostToolUse.
	if toolName == "Write" || toolName == "Edit" {
		return nil
	}

	runID, err := resolveOrCreateRun(s, sessionID, cwd)
	if err != nil {
		return err
	}

	agentID := s.ActiveAgentID(runID)
	payload := instrumentation.ToolUsePayload{AgentID: agentID, ToolName: toolName, ToolInput: toolInput}
	ev, err := instrumentation.NewEvent(runID, instrumentation.EventToolUse, payload)
	if err != nil {
		return err
	}
	return s.WriteEvent(ev)
}

func handleStopFailure(s *store.SQLiteStore, sessionID, reason string) error {
	runID, err := s.ResolveRunBySession(sessionID)
	if err != nil {
		return err
	}
	if runID == "" {
		return nil
	}

	payload := instrumentation.RunErrorPayload{ErrorReason: reason}
	ev, err := instrumentation.NewEvent(runID, instrumentation.EventRunError, payload)
	if err != nil {
		return err
	}
	return s.WriteEvent(ev)
}

func handleTaskCreated(s *store.SQLiteStore, sessionID, taskID, title, cwd string) error {
	runID, err := resolveOrCreateRun(s, sessionID, cwd)
	if err != nil {
		return err
	}

	payload := instrumentation.TaskCreatedPayload{TaskID: taskID, Title: title}
	ev, err := instrumentation.NewEvent(runID, instrumentation.EventTaskCreated, payload)
	if err != nil {
		return err
	}
	return s.WriteEvent(ev)
}

func handleTaskCompleted(s *store.SQLiteStore, sessionID, taskID string) error {
	runID, err := s.ResolveRunBySession(sessionID)
	if err != nil {
		return err
	}
	if runID == "" {
		return nil
	}

	payload := instrumentation.TaskCompletedPayload{TaskID: taskID}
	ev, err := instrumentation.NewEvent(runID, instrumentation.EventTaskCompleted, payload)
	if err != nil {
		return err
	}
	return s.WriteEvent(ev)
}

func handlePostCompact(s *store.SQLiteStore, sessionID, compactType string) error {
	runID, err := s.ResolveRunBySession(sessionID)
	if err != nil {
		return err
	}
	if runID == "" {
		return nil
	}

	if compactType == "" {
		compactType = "auto"
	}

	payload := instrumentation.CompactionPayload{Type: compactType}
	ev, err := instrumentation.NewEvent(runID, instrumentation.EventCompaction, payload)
	if err != nil {
		return err
	}
	return s.WriteEvent(ev)
}

func handlePermissionDenied(s *store.SQLiteStore, sessionID, toolName string) error {
	runID, err := s.ResolveRunBySession(sessionID)
	if err != nil {
		return err
	}
	if runID == "" {
		return nil
	}

	payload := instrumentation.PermissionDeniedPayload{ToolName: toolName}
	ev, err := instrumentation.NewEvent(runID, instrumentation.EventPermissionDenied, payload)
	if err != nil {
		return err
	}
	return s.WriteEvent(ev)
}

// resolveOrCreateRun finds the run for a session_id, or creates a new run
// if none exists (lazy creation — the run is only born on first real activity).
func resolveOrCreateRun(s *store.SQLiteStore, sessionID, cwd string) (string, error) {
	runID, err := s.ResolveRunBySession(sessionID)
	if err != nil {
		return "", err
	}
	if runID != "" {
		return runID, nil
	}
	return createRunForSession(s, sessionID, cwd)
}
