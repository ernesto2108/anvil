//go:build dashboard

package dashboard

import (
	"context"
	"fmt"
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
}

// --- Bindings ---------------------------------------------------------------

// GetRuns retorna una lista paginada de runs que coinciden con la consulta dada.
func (a *App) GetRuns(q RunsQuery) ([]RunDTO, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	rows, err := a.store.ListRuns(ctx, q.Limit, q.Offset)
	if err != nil {
		return nil, err
	}
	out := make([]RunDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toRunDTO(r))
	}
	return out, nil
}

// GetRun retorna el detalle de un run identificado por su ID.
func (a *App) GetRun(_ string) (*RunDetailDTO, error) {
	return nil, nil
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

// GetMetrics retorna los agregados de los últimos 30 runs terminados.
// No toma parámetros en MVP (ver D7). Retorna (nil, err) si el store falla.
func (a *App) GetMetrics() (*MetricsDTO, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	m, err := a.store.GetMetrics(ctx)
	if err != nil {
		return nil, err
	}
	dto := toMetricsDTO(m)
	return &dto, nil
}

// GetFiles retorna archivos producidos o consumidos por un agente dentro de un run.
func (a *App) GetFiles(_, _ string) ([]FileDTO, error) {
	return []FileDTO{}, nil
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
				Tokens:     r.TokensTotal,
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
		ID:          d.AgentID,
		Name:        d.AgentRole,
		Status:      d.Status,
		DurationMs:  d.DurationMs,
		TokensIn:    d.TokensInput,
		TokensOut:   d.TokensOutput,
		TokensTotal: d.TokensTotal,
		ErrorMsg:    d.ErrorMsg,
	}
	if d.StartedAt != nil {
		agent.StartedAt = d.StartedAt.Format(time.RFC3339Nano)
	}
	if d.EndedAt != nil {
		agent.EndedAt = d.EndedAt.Format(time.RFC3339Nano)
	}

	fileDTOs := make([]FileDTO, 0, len(files))
	for _, f := range files {
		fileDTOs = append(fileDTOs, FileDTO{Path: f.Path, Action: f.Operation})
	}

	return &AgentDetailDTO{
		Agent:  agent,
		Files:  fileDTOs,
		Output: "", // placeholder — captura real pendiente de DASH-FEAT-013
	}
}

// toMetricsDTO convierte un store.Metrics al DTO expuesto al frontend.
func toMetricsDTO(m store.Metrics) MetricsDTO {
	tokensPerRun := make([]TokensPerRunPoint, 0, len(m.TokensPerRun))
	for _, r := range m.TokensPerRun {
		tokensPerRun = append(tokensPerRun, TokensPerRunPoint{
			RunID:     r.RunID,
			StartedAt: r.StartedAt.Format(time.RFC3339Nano),
			Tokens:    r.Tokens,
		})
	}

	avgTimePerAgent := make([]AgentTimePoint, 0, len(m.AvgTimePerAgent))
	for _, a := range m.AvgTimePerAgent {
		avgTimePerAgent = append(avgTimePerAgent, AgentTimePoint{
			Role:          a.Role,
			AvgDurationMs: a.AvgDurationMs,
			RunsIncluded:  a.RunsIncluded,
		})
	}

	return MetricsDTO{
		RunsCount:           m.RunsCount,
		HasEnoughData:       m.HasEnoughData,
		TotalTokens:         m.TotalTokens,
		TotalTokensDeltaPct: m.TotalTokensDeltaPct,
		AvgDurationMs:       m.AvgDurationMs,
		AvgDurationDeltaMs:  m.AvgDurationDeltaMs,
		SuccessRate:         m.SuccessRate,
		SuccessRateDelta:    m.SuccessRateDelta,
		TokensPerRun:        tokensPerRun,
		AvgTimePerAgent:     avgTimePerAgent,
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
		ID:          r.ID,
		TaskID:      r.TaskID,
		TaskDesc:    r.TaskDesc,
		Status:      r.Status,
		Complexity:  r.Complexity,
		Provider:    r.Provider,
		StartedAt:   r.StartedAt.Format(time.RFC3339Nano),
		EndedAt:     endedAt,
		DurationMs:  durationMs,
		TotalTokens: r.TotalTokens,
		FilesCount:  r.FilesCount,
		AgentsCount: r.AgentsCount,
		QAScore:     r.QAScore,
	}
}
