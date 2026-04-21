package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/imports"
	"github.com/ernesto2108/anvil/internal/dashboard/writer"
	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// analyzerOnce guards singleton creation for the import Analyzer.
var (
	analyzerOnce sync.Once
	analyzer     *imports.Analyzer
)

// getAnalyzer returns the package-level Analyzer, creating it once with cwd as root.
func getAnalyzer(cwd string) *imports.Analyzer {
	analyzerOnce.Do(func() {
		analyzer = imports.New(cwd)
	})
	return analyzer
}

// analyzableExts is the set of file extensions supported by the import analyzer.
var analyzableExts = map[string]bool{
	".go":  true,
	".ts":  true,
	".tsx": true,
	".js":  true,
	".jsx": true,
	".rs":  true,
	".py":  true,
}

// parentRunID returns the run ID exported by the orchestrator when the
// current claude subprocess was launched as part of an `anvil run` pipeline.
// An empty string means this hook is from a standalone Claude Code session.
func parentRunID() string {
	return os.Getenv("ANVIL_PARENT_RUN_ID")
}

// envAgentID returns the agent ID exported by the orchestrator for the
// claude subprocess that produced the current hook event. Empty when the
// hook is not part of an anvil-managed pipeline.
func envAgentID() string {
	return os.Getenv("ANVIL_AGENT_ID")
}

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
func translateHook(raw []byte, w *writer.EventWriter) error {
	var env hookEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("parse hook JSON: %w", err)
	}

	if env.SessionID == "" {
		return fmt.Errorf("hook JSON missing session_id")
	}

	switch env.HookEventName {
	case "SessionStart":
		return handleSessionStart(w, env.SessionID, env.CWD)
	case "UserPromptSubmit":
		return handleUserPromptSubmit(w, env.SessionID, env.Prompt, env.CWD)
	case "SubagentStart":
		return handleSubagentStart(w, env.SessionID, env.AgentID, env.AgentType, env.CWD)
	case "SubagentStop":
		return handleSubagentStop(w, env.SessionID, env.AgentID, env.LastAssistantMessage)
	case "PreToolUse":
		return handlePreToolUse(w, env.SessionID, env.ToolName, env.ToolInput, env.CWD)
	case "PostToolUse":
		return handlePostToolUse(w, env.SessionID, env.ToolName, env.ToolInput, env.CWD)
	case "StopFailure":
		return handleStopFailure(w, env.SessionID, env.StopFailureReason)
	case "TaskCreated":
		return handleTaskCreated(w, env.SessionID, env.TaskID, env.TaskTitle, env.CWD)
	case "TaskCompleted":
		return handleTaskCompleted(w, env.SessionID, env.TaskID)
	case "PostCompact":
		return handlePostCompact(w, env.SessionID, env.CompactType)
	case "PermissionDenied":
		return handlePermissionDenied(w, env.SessionID, env.ToolName)
	case "Stop":
		return handleStop(w, env.SessionID, env.LastAssistantMessage)
	case "SessionEnd":
		return handleSessionEnd(w, env.SessionID)
	default:
		return nil
	}
}

// handleSessionStart creates a new run and registers a synthetic "direct"
// agent so that standalone Claude Code sessions appear in the dashboard with
// agents_count > 0, just like anvil-orchestrated pipeline runs.
//
// This only fires for top-level sessions (no ANVIL_PARENT_RUN_ID). Subprocesses
// launched by `anvil run` already have their run/agent lifecycle managed by the
// orchestrator.
func handleSessionStart(w *writer.EventWriter, sessionID, cwd string) error {
	if parentRunID() != "" {
		return nil
	}

	runID, err := createRunForSession(w, sessionID, cwd)
	if err != nil {
		return err
	}

	agentID := fmt.Sprintf("direct-%s", sessionID[:8])
	payload := instrumentation.AgentStartPayload{
		AgentID:   agentID,
		AgentRole: "direct",
	}
	ev, err := instrumentation.NewEvent(runID, instrumentation.EventAgentStart, payload)
	if err != nil {
		return err
	}
	return w.WriteEvent(ev)
}

// createRunForSession creates a new run linked to the given session.
func createRunForSession(w *writer.EventWriter, sessionID, cwd string) (string, error) {
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
	if err := w.WriteEvent(ev); err != nil {
		return "", err
	}
	return runID, nil
}

// projectFromCWD extracts the project name from a working directory path.
func projectFromCWD(cwd string) string {
	if cwd == "" {
		return ""
	}
	for len(cwd) > 1 && cwd[len(cwd)-1] == '/' {
		cwd = cwd[:len(cwd)-1]
	}
	for i := len(cwd) - 1; i >= 0; i-- {
		if cwd[i] == '/' {
			return cwd[i+1:]
		}
	}
	return cwd
}

// branchFromCWD returns the current git branch for the given working directory.
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

func handleUserPromptSubmit(w *writer.EventWriter, sessionID, prompt, cwd string) error {
	runID, err := resolveOrCreateRun(w, sessionID, cwd)
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
	if err := w.WriteEvent(ev); err != nil {
		return err
	}

	// When part of an anvil-orchestrated run, the parent already owns the
	// task description (the user's real objective). The per-agent prompt
	// would clobber it with internal scaffolding ("Complexity: Small …").
	if parentRunID() != "" {
		return nil
	}
	return w.UpdateTaskDesc(runID, prompt)
}

func handleSubagentStart(w *writer.EventWriter, sessionID, agentID, agentType, cwd string) error {
	runID, err := resolveOrCreateRun(w, sessionID, cwd)
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
	return w.WriteEvent(ev)
}

func handleSubagentStop(w *writer.EventWriter, sessionID, agentID, lastMessage string) error {
	runID, err := resolveOrCreateRun(w, sessionID, "")
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

	var durationMs int64
	if startedAt, ok := w.AgentStartedAt(runID, agentID); ok {
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
	return w.WriteEvent(ev)
}

func handlePostToolUse(w *writer.EventWriter, sessionID, toolName string, toolInput json.RawMessage, cwd string) error {
	if toolName != "Write" && toolName != "Edit" {
		return nil
	}

	runID, err := resolveOrCreateRun(w, sessionID, cwd)
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

	agentID := envAgentID()
	if agentID == "" {
		agentID = w.ActiveAgentID(runID)
	}

	var diff string
	const maxDiffBytes = 50 * 1024
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
	if err := w.WriteEvent(ev); err != nil {
		return err
	}

	// Best-effort: detect import edges and persist them.
	// Errors are logged and never propagate to the caller.
	analyzeFileEdges(w, runID, ti.FilePath, cwd)
	return nil
}

// analyzeFileEdges runs the import analyzer on filePath and persists any edges
// found. It is always best-effort: all errors are logged, never returned.
func analyzeFileEdges(w *writer.EventWriter, runID, filePath, cwd string) {
	if !analyzableExts[strings.ToLower(filepath.Ext(filePath))] {
		return
	}
	if cwd == "" {
		return
	}

	an := getAnalyzer(cwd)

	touchedPaths, err := w.TouchedPathsForRun(runID)
	if err != nil {
		log.Printf("anvil/emit: TouchedPathsForRun: %v", err)
		return
	}

	edges, err := an.Analyze(filePath, touchedPaths)
	if err != nil {
		log.Printf("anvil/emit: Analyze(%s): %v", filePath, err)
		return
	}

	if len(edges) == 0 {
		return
	}

	if err := w.WriteFileEdges(runID, edges); err != nil {
		log.Printf("anvil/emit: WriteFileEdges: %v", err)
	}
}

// handleStop fires at the end of each Claude turn. For direct sessions it
// appends the assistant's response to the "direct" agent output so the
// dashboard shows what Claude actually said/did across all turns.
//
// Skipped for anvil-orchestrated subprocesses — the orchestrator captures
// output via SubagentStop.
func handleStop(w *writer.EventWriter, sessionID, lastMessage string) error {
	if parentRunID() != "" || lastMessage == "" {
		return nil
	}

	runID, err := w.ResolveRunBySession(sessionID)
	if err != nil || runID == "" {
		return err
	}

	agentID := fmt.Sprintf("direct-%s", sessionID[:8])
	return w.AppendAgentOutput(runID, agentID, lastMessage)
}

func handleSessionEnd(w *writer.EventWriter, sessionID string) error {
	// When this hook fires from a claude subprocess launched by `anvil run`,
	// the orchestrator owns the run lifecycle and emits its own run.end.
	// Closing the run here would mark it finished while sibling agents are
	// still in flight.
	if parentRunID() != "" {
		return nil
	}

	runID, err := w.ResolveRunBySession(sessionID)
	if err != nil {
		return err
	}
	if runID == "" {
		return nil
	}

	// Close the synthetic "direct" agent opened by handleSessionStart.
	agentID := fmt.Sprintf("direct-%s", sessionID[:8])
	var durationMs int64
	if startedAt, ok := w.AgentStartedAt(runID, agentID); ok {
		if t, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", startedAt); err == nil {
			durationMs = time.Since(t).Milliseconds()
		}
	}
	agentEnd := instrumentation.AgentEndPayload{
		AgentID:    agentID,
		Status:     "success",
		DurationMs: durationMs,
	}
	if ev, err := instrumentation.NewEvent(runID, instrumentation.EventAgentEnd, agentEnd); err == nil {
		_ = w.WriteEvent(ev)
	}

	payload := instrumentation.RunEndPayload{
		Status: "success",
	}

	ev, err := instrumentation.NewEvent(runID, instrumentation.EventRunEnd, payload)
	if err != nil {
		return err
	}
	if err := w.WriteEvent(ev); err != nil {
		return err
	}

	return w.ComputeRunTotals(runID)
}

func handlePreToolUse(w *writer.EventWriter, sessionID, toolName string, toolInput json.RawMessage, cwd string) error {
	if toolName == "Write" || toolName == "Edit" {
		return nil
	}

	runID, err := resolveOrCreateRun(w, sessionID, cwd)
	if err != nil {
		return err
	}

	agentID := envAgentID()
	if agentID == "" {
		agentID = w.ActiveAgentID(runID)
	}

	source, mcpServer := classifyTool(toolName)
	payload := instrumentation.ToolUsePayload{
		AgentID:   agentID,
		ToolName:  toolName,
		ToolInput: toolInput,
		Source:    source,
		MCPServer: mcpServer,
	}
	ev, err := instrumentation.NewEvent(runID, instrumentation.EventToolUse, payload)
	if err != nil {
		return err
	}
	return w.WriteEvent(ev)
}

// classifyTool returns the source ("native" or "mcp") and the MCP server name
// for a given tool name. MCP tools follow the pattern mcp__<server>__<method>.
func classifyTool(toolName string) (source, mcpServer string) {
	parts := strings.SplitN(toolName, "__", 3)
	if len(parts) == 3 && parts[0] == "mcp" {
		return "mcp", parts[1]
	}
	return "native", ""
}

func handleStopFailure(w *writer.EventWriter, sessionID, reason string) error {
	runID, err := w.ResolveRunBySession(sessionID)
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
	return w.WriteEvent(ev)
}

func handleTaskCreated(w *writer.EventWriter, sessionID, taskID, title, cwd string) error {
	runID, err := resolveOrCreateRun(w, sessionID, cwd)
	if err != nil {
		return err
	}

	payload := instrumentation.TaskCreatedPayload{TaskID: taskID, Title: title}
	ev, err := instrumentation.NewEvent(runID, instrumentation.EventTaskCreated, payload)
	if err != nil {
		return err
	}
	return w.WriteEvent(ev)
}

func handleTaskCompleted(w *writer.EventWriter, sessionID, taskID string) error {
	runID, err := w.ResolveRunBySession(sessionID)
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
	return w.WriteEvent(ev)
}

func handlePostCompact(w *writer.EventWriter, sessionID, compactType string) error {
	runID, err := w.ResolveRunBySession(sessionID)
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
	return w.WriteEvent(ev)
}

func handlePermissionDenied(w *writer.EventWriter, sessionID, toolName string) error {
	runID, err := w.ResolveRunBySession(sessionID)
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
	return w.WriteEvent(ev)
}

// resolveOrCreateRun finds the run for a session_id, or creates a new run
// if none exists (lazy creation — the run is only born on first real activity).
//
// When ANVIL_PARENT_RUN_ID is set the lookup is bypassed entirely: the hook
// belongs to a claude subprocess that the anvil orchestrator already
// registered, so events must attach to that run rather than spawning an
// orphan sibling keyed by the subprocess's session_id.
func resolveOrCreateRun(w *writer.EventWriter, sessionID, cwd string) (string, error) {
	if rid := parentRunID(); rid != "" {
		return rid, nil
	}
	runID, err := w.ResolveRunBySession(sessionID)
	if err != nil {
		return "", err
	}
	if runID != "" {
		return runID, nil
	}
	return createRunForSession(w, sessionID, cwd)
}
