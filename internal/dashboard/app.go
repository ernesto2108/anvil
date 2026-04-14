//go:build dashboard

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

	"github.com/ernesto2108/anvil/internal/dashboard/store"
)

// App contiene el contexto de ejecución de Wails y el store de solo lectura
// usado por los bindings expuestos al frontend.
type App struct {
	ctx   context.Context
	store Store
}

// NewApp conecta la dependencia de store. Store es obligatorio: pasar nil
// solo en tests que no ejerciten bindings.
func NewApp(s Store) *App {
	return &App{store: s}
}

// Startup es invocado por Wails una vez que la ventana nativa está lista.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	go a.cleanupLoop(ctx)
}

// cleanupLoop marks stale runs as abandoned every 5 minutes.
// Stops when the Wails context is cancelled (app shutdown).
func (a *App) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if cleaned, err := a.store.CleanupStaleRuns(2); err != nil {
				log.Printf("dashboard: periodic cleanup: %v", err)
			} else if cleaned > 0 {
				log.Printf("dashboard: marked %d stale runs as abandoned", cleaned)
			}
		}
	}
}

// --- Bindings ---------------------------------------------------------------

// GetProjects returns distinct project names for the project filter.
func (a *App) GetProjects() ([]string, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.store.ListProjects(ctx)
}

// GetRuns retorna una lista paginada de runs que coinciden con la consulta dada.
func (a *App) GetRuns(q RunsQuery) ([]RunDTO, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	// Normalizar fechas YYYY-MM-DD del frontend a RFC3339 para comparación con started_at.
	startDate := normalizeDate(q.StartDate, false)
	endDate := normalizeDate(q.EndDate, true)

	rows, err := a.store.ListRuns(ctx, q.Limit, q.Offset, q.Status, startDate, endDate, q.Project)
	if err != nil {
		return nil, err
	}
	out := make([]RunDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toRunDTO(r))
	}
	return out, nil
}

// GetRunSummary retorna el resumen de un run identificado por su ID.
// Retorna (nil, nil) si el run no existe — NO es un error.
func (a *App) GetRunSummary(runID string) (*RunDTO, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	r, err := a.store.GetRunSummary(ctx, runID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	dto := toRunDTO(*r)
	return &dto, nil
}

// GetAgents retorna todos los agentes pertenecientes al run identificado por runID.
func (a *App) GetAgents(_ string) ([]AgentDTO, error) {
	return []AgentDTO{}, nil
}

// GetAgent retorna el detalle de un agente individual dentro de un run.
// Retorna (nil, nil) si el agente no existe.
func (a *App) GetAgent(runID, agentID string) (*AgentDetailDTO, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	detail, files, err := a.store.GetAgentDetail(ctx, runID, agentID)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, nil
	}
	return toAgentDetailDTO(detail, files), nil
}

// GetChildRuns returns child runs for a parent run (cross-service orchestration).
func (a *App) GetChildRuns(parentRunID string) ([]RunDTO, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	rows, err := a.store.ListChildRuns(ctx, parentRunID)
	if err != nil {
		return nil, err
	}
	out := make([]RunDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toRunDTO(r))
	}
	return out, nil
}

// GetFlow retorna el grafo de flujo de ejecución para un run.
func (a *App) GetFlow(runID string) (*FlowDTO, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	rows, err := a.store.ListAgentsByRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	return toFlowDTO(runID, rows), nil
}

// GetSessionDetail returns run summary + files changed + agent outputs
// for the session detail view.
func (a *App) GetSessionDetail(runID string) (*SessionDetailDTO, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	// Run summary
	r, err := a.store.GetRunSummary(ctx, runID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}

	// Files
	fileRows, err := a.store.ListFilesByRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	files := make([]FileDTO, 0, len(fileRows))
	for _, f := range fileRows {
		files = append(files, FileDTO{Path: f.Path, Action: f.Operation, Diff: f.Diff, AgentID: f.AgentID})
	}

	// Agents (reuse existing query for roles/status, then fetch output per agent)
	agentRows, err := a.store.ListAgentsByRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	agents := make([]SessionAgentDTO, 0, len(agentRows))
	for _, ag := range agentRows {
		sa := SessionAgentDTO{
			ID:         ag.AgentID,
			Role:       ag.AgentRole,
			Status:     ag.Status,
			DurationMs: ag.DurationMs,
		}
		if ag.StartedAt != nil {
			sa.StartedAt = ag.StartedAt.Format(time.RFC3339Nano)
		}
		if ag.EndedAt != nil {
			sa.EndedAt = ag.EndedAt.Format(time.RFC3339Nano)
		}
		// Fetch output from agent detail
		detail, _, detailErr := a.store.GetAgentDetail(ctx, runID, ag.AgentID)
		if detailErr == nil && detail != nil {
			sa.Output = detail.Output
		}
		agents = append(agents, sa)
	}

	// Activity events for trace timeline (with agent attribution)
	rawEvents, _ := a.store.ListActivityEvents(ctx, runID)
	activityEvents := make([]ActivityEventDTO, 0, len(rawEvents))
	for _, ev := range rawEvents {
		activityEvents = append(activityEvents, ActivityEventDTO{
			Timestamp: ev.Timestamp,
			AgentID:   ev.AgentID,
		})
	}

	runDTO := toRunDTO(*r)

	// If branch is empty (pre-migration runs), resolve it live from the project dir.
	if runDTO.Branch == "" && runDTO.Project != "" {
		runDTO.Branch = liveBranch(runDTO.Project)
	}

	// Tool usage breakdown
	toolRows, _ := a.store.ListToolUsageByRun(ctx, runID)
	toolUsage := make([]ToolUsageDTO, 0, len(toolRows))
	for _, t := range toolRows {
		toolUsage = append(toolUsage, ToolUsageDTO{ToolName: t.ToolName, Count: t.Count})
	}

	// Enrich RunDTO with counts
	toolTotal, _ := a.store.TotalToolUsesByRun(ctx, runID)
	runDTO.ToolUsesCount = toolTotal
	compactions, _ := a.store.CountCompactions(ctx, runID)
	runDTO.CompactionsCount = compactions

	// Tool use details (individual commands)
	detailRows, _ := a.store.ListToolUseDetailsByRun(ctx, runID)
	toolDetails := make([]ToolUseDetailDTO, 0, len(detailRows))
	for _, d := range detailRows {
		cmd := extractCommand(d.ToolName, d.ToolInput)
		if cmd != "" {
			toolDetails = append(toolDetails, ToolUseDetailDTO{
				ToolName:  d.ToolName,
				Command:   cmd,
				AgentID:   d.AgentID,
				Timestamp: d.Timestamp,
			})
		}
	}

	// Tasks
	taskRows, _ := a.store.ListTasksByRun(ctx, runID)
	tasks := make([]TaskDTO, 0, len(taskRows))
	for _, t := range taskRows {
		tasks = append(tasks, TaskDTO{
			ID:          t.TaskID,
			Title:       t.Title,
			Status:      t.Status,
			CreatedAt:   t.CreatedAt,
			CompletedAt: t.CompletedAt,
		})
	}

	dto := &SessionDetailDTO{
		Run:            runDTO,
		Files:          files,
		Agents:         agents,
		ActivityEvents: activityEvents,
		ToolUsage:      toolUsage,
		ToolDetails:    toolDetails,
		Tasks:          tasks,
	}
	return dto, nil
}

// GetToolUsage returns tool usage breakdown for a run.
func (a *App) GetToolUsage(runID string) ([]ToolUsageDTO, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	rows, err := a.store.ListToolUsageByRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	out := make([]ToolUsageDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, ToolUsageDTO{ToolName: r.ToolName, Count: r.Count})
	}
	return out, nil
}

// GetPrompts returns all prompts for a run, ordered by sequence ASC.
func (a *App) GetPrompts(runID string) ([]PromptDTO, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	rows, err := a.store.ListPromptsByRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	out := make([]PromptDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, PromptDTO{
			Sequence:  r.Sequence,
			Prompt:    r.Prompt,
			Timestamp: r.Timestamp,
		})
	}
	return out, nil
}

// GetTurnStats returns activity statistics per turn for a run.
func (a *App) GetTurnStats(runID string) ([]TurnStatsDTO, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	rows, err := a.store.GetTurnStats(ctx, runID)
	if err != nil {
		return nil, err
	}
	out := make([]TurnStatsDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, TurnStatsDTO{
			TurnNumber:    r.TurnNumber,
			Prompt:        r.Prompt,
			Timestamp:     r.Timestamp,
			EndTimestamp:  r.EndTimestamp,
			FilesCount:    r.FilesCount,
			ToolUsesCount: r.ToolUsesCount,
			AgentsCount:   r.AgentsCount,
		})
	}
	return out, nil
}

// GetTasks returns all tasks for a run.
func (a *App) GetTasks(runID string) ([]TaskDTO, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	rows, err := a.store.ListTasksByRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	out := make([]TaskDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, TaskDTO{
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
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.store.DeleteRun(ctx, runID)
}

// DeleteRuns deletes multiple runs in a single transaction.
func (a *App) DeleteRuns(runIDs []string) error {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.store.DeleteRuns(ctx, runIDs)
}

// RevertFile applies the reverse of the given diff to restore a file.
// It uses git apply -R in the project directory ~/projects/<project>.
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

	// Write diff to temp file for git apply
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

// liveBranch runs `git branch --show-current` in ~/projects/<project>.
// Returns "" silently on any failure.
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
	// trim newline
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

// extractCommand extracts a human-readable command string from a tool_input JSON.
// For Bash, returns the "command" field. For other tools, returns a short summary.
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

// --- conversores privados ----------------------------------------------------

// toFlowDTO construye un FlowDTO a partir de las filas de agentes de un run.
//
// Lógica de aristas:
//  1. Si el agente tiene depends_on no vacío, se crea una arista por cada
//     ID listado (formato JSON array: ["agent-id-1"]).
//  2. Si depends_on está vacío, se usa fallback secuencial: el agente
//     anterior en el slice conecta con el actual.
//  3. Las aristas duplicadas (cuando depends_on y secuencia coinciden) se evitan
//     con un set de IDs ya emitidos.
func toFlowDTO(runID string, rows []store.AgentRow) *FlowDTO {
	nodes := make([]FlowNode, 0, len(rows))
	edges := make([]FlowEdge, 0, len(rows))
	edgeSeen := make(map[string]struct{}, len(rows))

	for _, r := range rows {
		node := FlowNode{
			ID:   r.AgentID,
			Type: "agentNode",
			Data: FlowNodeData{
				Label:      r.AgentRole,
				Status:     r.Status,
				DurationMs: r.DurationMs,
			},
		}
		nodes = append(nodes, node)
	}

	for i, r := range rows {
		// Parsear depends_on (JSON array de strings almacenado como TEXT).
		deps := parseDependsOn(r.DependsOn)

		if len(deps) > 0 {
			for _, depID := range deps {
				key := depID + "->" + r.AgentID
				if _, already := edgeSeen[key]; already {
					continue
				}
				edgeSeen[key] = struct{}{}
				edges = append(edges, FlowEdge{
					ID:     fmt.Sprintf("e-%s-%s", depID, r.AgentID),
					Source: depID,
					Target: r.AgentID,
				})
			}
		} else if i > 0 {
			// Fallback secuencial: conectar con el agente anterior.
			prev := rows[i-1].AgentID
			key := prev + "->" + r.AgentID
			if _, already := edgeSeen[key]; !already {
				edgeSeen[key] = struct{}{}
				edges = append(edges, FlowEdge{
					ID:     fmt.Sprintf("e-%s-%s", prev, r.AgentID),
					Source: prev,
					Target: r.AgentID,
				})
			}
		}
	}

	return &FlowDTO{Nodes: nodes, Edges: edges}
}

// parseDependsOn extrae los IDs de dependencias del campo TEXT de SQLite.
// El campo puede ser un JSON array (p.ej. '["agent-a","agent-b"]') o vacío/null.
// Retorna nil si no hay dependencias válidas.
func parseDependsOn(raw string) []string {
	if raw == "" || raw == "[]" || raw == "null" {
		return nil
	}
	// Decodificación manual liviana: strip corchetes y dividir por coma.
	// Evita importar encoding/json para un campo simple en un conversor crítico.
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

// splitJSON divide un string de elementos JSON separados por coma,
// respetando strings con comillas (no soporta anidado).
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

// trimQuotes elimina comillas dobles y espacios de un token JSON string.
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

// toAgentDetailDTO convierte los datos crudos del store al DTO expuesto al frontend.
func toAgentDetailDTO(d *store.AgentDetail, files []store.FileRow) *AgentDetailDTO {
	agent := AgentDTO{
		ID:         d.AgentID,
		Name:       d.AgentRole,
		Status:     d.Status,
		DurationMs: d.DurationMs,
		ErrorMsg:   d.ErrorMsg,
	}
	if d.StartedAt != nil {
		agent.StartedAt = d.StartedAt.Format(time.RFC3339Nano)
	}
	if d.EndedAt != nil {
		agent.EndedAt = d.EndedAt.Format(time.RFC3339Nano)
	}

	fileDTOs := make([]FileDTO, 0, len(files))
	for _, f := range files {
		fileDTOs = append(fileDTOs, FileDTO{Path: f.Path, Action: f.Operation, Diff: f.Diff, AgentID: f.AgentID})
	}

	return &AgentDetailDTO{
		Agent:  agent,
		Files:  fileDTOs,
		Output: d.Output,
	}
}

// toRunDTO convierte un RunSummary del store al DTO expuesto al frontend.
func toRunDTO(r store.RunSummary) RunDTO {
	endedAt := ""
	if r.EndedAt != nil {
		endedAt = r.EndedAt.Format(time.RFC3339Nano)
	}

	var durationMs int64
	if r.DurationMs != nil {
		durationMs = *r.DurationMs
	}

	return RunDTO{
		ID:            r.ID,
		TaskID:        r.TaskID,
		TaskDesc:      r.TaskDesc,
		Status:        r.Status,
		Complexity:    r.Complexity,
		Provider:      r.Provider,
		Project:       r.Project,
		StartedAt:     r.StartedAt.Format(time.RFC3339Nano),
		EndedAt:       endedAt,
		DurationMs:    durationMs,
		FilesCount:    r.FilesCount,
		AgentsCount:   r.AgentsCount,
		ParentRunID:   r.ParentRunID,
		ChildrenCount: r.ChildrenCount,
		Branch:        r.Branch,
		ErrorReason:   r.ErrorReason,
	}
}

// normalizeDate convierte una fecha YYYY-MM-DD del frontend a RFC3339 en zona local.
// Si endOfDay es true, usa 23:59:59.999999999 para incluir todo el día.
// Retorna cadena vacía si raw está vacío o no es una fecha válida.
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
