//go:build dashboard

// Deprecated: This Wails-based dashboard will be replaced by a standalone
// Rust + Tauri 2 application (anvil-dashboard). The Go layers (entity, dto,
// mapper, query, writer) are designed to be reusable; only the Wails bindings
// and frontend embedding in this file will be removed once Tauri reaches parity.
package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/dto"
	"github.com/ernesto2108/anvil/internal/dashboard/mapper"
)

// App contiene el contexto de ejecución de Wails y las dependencias de lectura/escritura
// usadas por los bindings expuestos al frontend.
type App struct {
	ctx    context.Context
	reader DashboardReader
	writer DashboardWriter
}

// NewApp conecta las dependencias de lectura y escritura.
func NewApp(reader DashboardReader, writer DashboardWriter) *App {
	return &App{reader: reader, writer: writer}
}

// Startup es invocado por Wails una vez que la ventana nativa está lista.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	go a.cleanupLoop(ctx)
}

// cleanupLoop marks stale runs as abandoned every 30 seconds.
func (a *App) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if cleaned, err := a.writer.CleanupStaleRuns(2); err != nil {
				log.Printf("dashboard: periodic cleanup: %v", err)
			} else if cleaned > 0 {
				log.Printf("dashboard: marked %d stale runs as abandoned", cleaned)
			}
		}
	}
}

func (a *App) getCtx() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}

// --- Bindings ---------------------------------------------------------------

// GetProjects returns distinct project names for the project filter.
func (a *App) GetProjects() ([]string, error) {
	return a.reader.ListProjects(a.getCtx())
}

// GetRuns retorna una lista paginada de runs que coinciden con la consulta dada.
func (a *App) GetRuns(q dto.RunsQuery) ([]dto.RunDTO, error) {
	startDate := normalizeDate(q.StartDate, false)
	endDate := normalizeDate(q.EndDate, true)

	entities, err := a.reader.ListRuns(a.getCtx(), q.Limit, q.Offset, q.Status, startDate, endDate, q.Project)
	if err != nil {
		return nil, err
	}
	return mapper.ToRunDTOs(entities), nil
}

// GetRunSummary retorna el resumen de un run identificado por su ID.
func (a *App) GetRunSummary(runID string) (*dto.RunDTO, error) {
	r, err := a.reader.GetRunSummary(a.getCtx(), runID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	d := mapper.ToRunDTO(*r)
	return &d, nil
}

// GetAgents retorna todos los agentes pertenecientes al run identificado por runID.
func (a *App) GetAgents(_ string) ([]dto.AgentDTO, error) {
	return []dto.AgentDTO{}, nil
}

// GetAgent retorna el detalle de un agente individual dentro de un run.
func (a *App) GetAgent(runID, agentID string) (*dto.AgentDetailDTO, error) {
	detail, files, err := a.reader.GetAgentDetail(a.getCtx(), runID, agentID)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, nil
	}
	d := mapper.ToAgentDetailDTO(*detail, files)
	return &d, nil
}

// GetChildRuns returns child runs for a parent run (cross-service orchestration).
func (a *App) GetChildRuns(parentRunID string) ([]dto.RunDTO, error) {
	entities, err := a.reader.ListChildRuns(a.getCtx(), parentRunID)
	if err != nil {
		return nil, err
	}
	return mapper.ToRunDTOs(entities), nil
}

// GetFlow retorna el grafo de flujo de ejecución para un run.
func (a *App) GetFlow(runID string) (*dto.FlowDTO, error) {
	rows, err := a.reader.ListAgentsByRun(a.getCtx(), runID)
	if err != nil {
		return nil, err
	}
	d := mapper.ToFlowDTO(rows, parseDependsOn)
	return &d, nil
}

// GetSessionDetail returns run summary + files changed + agent outputs
// for the session detail view.
func (a *App) GetSessionDetail(runID string) (*dto.SessionDetailDTO, error) {
	ctx := a.getCtx()

	// Run summary
	r, err := a.reader.GetRunSummary(ctx, runID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}

	// Files
	fileRows, err := a.reader.ListFilesByRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	files := mapper.ToFileDTOs(fileRows)

	// Agents
	agentRows, err := a.reader.ListAgentsByRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	agents := make([]dto.SessionAgentDTO, 0, len(agentRows))
	for _, ag := range agentRows {
		var output string
		detail, _, detailErr := a.reader.GetAgentDetail(ctx, runID, ag.AgentID)
		if detailErr == nil && detail != nil {
			output = detail.Output
		}
		agents = append(agents, mapper.ToSessionAgentDTO(ag, output))
	}

	// Activity events
	rawEvents, _ := a.reader.ListActivityEvents(ctx, runID)
	activityEvents := make([]dto.ActivityEventDTO, 0, len(rawEvents))
	for _, ev := range rawEvents {
		activityEvents = append(activityEvents, dto.ActivityEventDTO{
			Timestamp: ev.Timestamp,
			AgentID:   ev.AgentID,
		})
	}

	runDTO := mapper.ToRunDTO(*r)

	// If branch is empty (pre-migration runs), resolve it live from the project dir.
	if runDTO.Branch == "" && runDTO.Project != "" {
		runDTO.Branch = liveBranch(runDTO.Project)
	}

	// Tool usage breakdown
	toolRows, _ := a.reader.ListToolUsageByRun(ctx, runID)
	toolUsage := mapper.ToToolUsageDTOs(toolRows)

	// Enrich RunDTO with counts
	toolTotal, _ := a.reader.TotalToolUsesByRun(ctx, runID)
	runDTO.ToolUsesCount = toolTotal
	compactions, _ := a.reader.CountCompactions(ctx, runID)
	runDTO.CompactionsCount = compactions

	// Tool use details (individual commands)
	detailRows, _ := a.reader.ListToolUseDetailsByRun(ctx, runID)
	toolDetails := make([]dto.ToolUseDetailDTO, 0, len(detailRows))
	for _, d := range detailRows {
		cmd := extractCommand(d.ToolName, d.ToolInput)
		if cmd != "" {
			toolDetails = append(toolDetails, dto.ToolUseDetailDTO{
				ToolName:  d.ToolName,
				Command:   cmd,
				AgentID:   d.AgentID,
				Timestamp: d.Timestamp,
			})
		}
	}

	// Tasks
	taskRows, _ := a.reader.ListTasksByRun(ctx, runID)
	tasks := make([]dto.TaskDTO, 0, len(taskRows))
	for _, t := range taskRows {
		tasks = append(tasks, dto.TaskDTO{
			ID:          t.TaskID,
			Title:       t.Title,
			Status:      t.Status,
			CreatedAt:   t.CreatedAt,
			CompletedAt: t.CompletedAt,
		})
	}

	return &dto.SessionDetailDTO{
		Run:            runDTO,
		Files:          files,
		Agents:         agents,
		ActivityEvents: activityEvents,
		ToolUsage:      toolUsage,
		ToolDetails:    toolDetails,
		Tasks:          tasks,
	}, nil
}

// GetToolUsage returns tool usage breakdown for a run.
func (a *App) GetToolUsage(runID string) ([]dto.ToolUsageDTO, error) {
	rows, err := a.reader.ListToolUsageByRun(a.getCtx(), runID)
	if err != nil {
		return nil, err
	}
	return mapper.ToToolUsageDTOs(rows), nil
}

// GetPrompts returns all prompts for a run, ordered by sequence ASC.
func (a *App) GetPrompts(runID string) ([]dto.PromptDTO, error) {
	rows, err := a.reader.ListPromptsByRun(a.getCtx(), runID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.PromptDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.PromptDTO{
			Sequence:  r.Sequence,
			Prompt:    r.Prompt,
			Timestamp: r.Timestamp,
		})
	}
	return out, nil
}

// GetTurnStats returns activity statistics per turn for a run.
func (a *App) GetTurnStats(runID string) ([]dto.TurnStatsDTO, error) {
	rows, err := a.reader.GetTurnStats(a.getCtx(), runID)
	if err != nil {
		return nil, err
	}
	return mapper.ToTurnStatsDTOs(rows), nil
}

// GetTasks returns all tasks for a run.
func (a *App) GetTasks(runID string) ([]dto.TaskDTO, error) {
	rows, err := a.reader.ListTasksByRun(a.getCtx(), runID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.TaskDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.TaskDTO{
			ID:          r.TaskID,
			Title:       r.Title,
			Status:      r.Status,
			CreatedAt:   r.CreatedAt,
			CompletedAt: r.CompletedAt,
		})
	}
	return out, nil
}

// DeleteRun deletes a single run and all associated data.
func (a *App) DeleteRun(runID string) error {
	return a.reader.DeleteRun(a.getCtx(), runID)
}

// DeleteRuns deletes multiple runs in a single transaction.
func (a *App) DeleteRuns(runIDs []string) error {
	return a.reader.DeleteRuns(a.getCtx(), runIDs)
}

// --- Error Bindings ---

// GetErrorGroups returns a filtered list of error groups.
func (a *App) GetErrorGroups(q dto.ErrorsQuery) ([]dto.ErrorGroupDTO, error) {
	rows, err := a.reader.ListErrorGroups(a.getCtx(), q.Status, q.Search)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ErrorGroupDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapper.ToErrorGroupDTO(r))
	}
	return out, nil
}

// GetErrorGroup returns the full detail of an error group.
func (a *App) GetErrorGroup(groupID string) (*dto.ErrorGroupDetailDTO, error) {
	ctx := a.getCtx()

	group, err := a.reader.GetErrorGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, nil
	}

	runs, _ := a.reader.ListErrorGroupRuns(ctx, groupID)
	history, _ := a.reader.ListErrorGroupHistory(ctx, groupID)
	trend, _ := a.reader.GetErrorTrend(ctx, groupID)

	d := mapper.ToErrorGroupDetailDTO(*group, runs, history, trend)
	return &d, nil
}

// UpdateErrorResolution updates the resolution status of an error group.
func (a *App) UpdateErrorResolution(req dto.UpdateResolutionRequest) error {
	return a.writer.UpdateErrorResolution(a.getCtx(), req.GroupID, req.Status, req.Notes, req.CommitLink)
}

// GetErrorsCounts returns summary counts of error groups by status.
func (a *App) GetErrorsCounts() (*dto.ErrorsCountDTO, error) {
	ec, err := a.reader.GetErrorCounts(a.getCtx())
	if err != nil {
		return nil, err
	}
	return &dto.ErrorsCountDTO{New: ec.New, Investigating: ec.Investigating, WeekTotal: ec.WeekTotal}, nil
}

// RevertFile applies the reverse of the given diff to restore a file.
func (a *App) RevertFile(project, filePath, diff string) error {
	if project == "" {
		return fmt.Errorf("proyecto no especificado")
	}
	if diff == "" {
		return fmt.Errorf("diff vacío — no hay cambios que revertir")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("no se pudo obtener el directorio home: %w", err)
	}

	projectDir := filepath.Join(home, "projects", project)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return fmt.Errorf("directorio del proyecto no encontrado: %s", projectDir)
	}

	tmp, err := os.CreateTemp("", "anvil-revert-*.patch")
	if err != nil {
		return fmt.Errorf("crear archivo temporal: %w", err)
	}
	defer os.Remove(tmp.Name()) //nolint:errcheck

	if _, err := tmp.WriteString(diff); err != nil {
		tmp.Close() //nolint:errcheck
		return fmt.Errorf("escribir diff temporal: %w", err)
	}
	tmp.Close() //nolint:errcheck

	cmd := exec.Command("git", "-C", projectDir, "apply", "-R", "--", tmp.Name())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git apply -R falló: %s: %w", string(out), err)
	}

	return nil
}

// --- Private helpers --------------------------------------------------------

// liveBranch runs `git branch --show-current` in ~/projects/<project>.
func liveBranch(project string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	dir := filepath.Join(home, "projects", project)
	out, err := exec.Command("git", "-C", dir, "branch", "--show-current").Output()
	if err != nil {
		return ""
	}
	s := string(out)
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

// extractCommand extracts a human-readable command string from a tool_input JSON.
func extractCommand(toolName, rawInput string) string {
	if rawInput == "" {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(rawInput), &m); err != nil {
		return ""
	}
	switch toolName {
	case "Bash":
		if cmd, ok := m["command"].(string); ok {
			return cmd
		}
	case "Grep":
		if pat, ok := m["pattern"].(string); ok {
			return "grep: " + pat
		}
	case "Glob":
		if pat, ok := m["pattern"].(string); ok {
			return "glob: " + pat
		}
	case "Read":
		if fp, ok := m["file_path"].(string); ok {
			return "read: " + fp
		}
	}
	return ""
}

// parseDependsOn extracts dependency IDs from the SQLite TEXT field.
func parseDependsOn(raw string) []string {
	if raw == "" || raw == "[]" || raw == "null" {
		return nil
	}
	s := raw
	if len(s) >= 2 && s[0] == '[' && s[len(s)-1] == ']' {
		s = s[1 : len(s)-1]
	}
	if s == "" {
		return nil
	}

	var result []string
	for _, part := range splitJSON(s) {
		part = trimQuotes(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// splitJSON splits comma-separated JSON elements respecting quoted strings.
func splitJSON(s string) []string {
	var parts []string
	inString := false
	start := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"':
			inString = !inString
		case ',':
			if !inString {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}

// trimQuotes removes double quotes and whitespace from a JSON string token.
func trimQuotes(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	return s
}

// normalizeDate converts a YYYY-MM-DD date from the frontend to RFC3339 in local timezone.
func normalizeDate(raw string, endOfDay bool) string {
	if raw == "" {
		return ""
	}
	t, err := time.ParseInLocation("2006-01-02", raw, time.Local)
	if err != nil {
		return ""
	}
	if endOfDay {
		t = t.Add(24*time.Hour - time.Nanosecond)
	}
	return t.UTC().Format(time.RFC3339Nano)
}
