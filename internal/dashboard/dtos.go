//go:build dashboard

package dashboard

// DTOs expuestos al frontend via Wails bindings.

// RunsQuery contiene los parámetros de filtro para listar runs.
type RunsQuery struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Status string `json:"status"` // reservado para filtro P1
}

// RunDTO es la representación resumida de un run individual.
type RunDTO struct {
	ID          string   `json:"id"`
	TaskID      string   `json:"taskId"`
	TaskDesc    string   `json:"taskDesc"`
	Status      string   `json:"status"`
	Complexity  string   `json:"complexity"`
	Provider    string   `json:"provider"`
	StartedAt   string   `json:"startedAt"`  // RFC3339Nano
	EndedAt     string   `json:"endedAt"`    // "" si todavía no terminó
	DurationMs  int64    `json:"durationMs"` // 0 si todavía no terminó
	TotalTokens int      `json:"totalTokens"`
	FilesCount  int      `json:"filesCount"`
	AgentsCount int      `json:"agentsCount"`
	QAScore     *float64 `json:"qaScore"` // nullable
}

// RunDetailDTO agrega un run con sus agentes.
type RunDetailDTO struct {
	Run    RunDTO     `json:"run"`
	Agents []AgentDTO `json:"agents"`
}

// AgentDTO es la representación completa de un agente dentro de un run.
type AgentDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`        // = agent_role
	Status      string `json:"status"`
	DurationMs  *int64 `json:"durationMs"`  // nil si el agente aún está corriendo
	TokensIn    *int   `json:"tokensIn"`
	TokensOut   *int   `json:"tokensOut"`
	TokensTotal *int   `json:"tokensTotal"`
	StartedAt   string `json:"startedAt"`   // RFC3339Nano, "" si no seteado
	EndedAt     string `json:"endedAt"`     // RFC3339Nano, "" si no terminó
	ErrorMsg    string `json:"errorMsg"`    // "" si no falló
}

// AgentDetailDTO agrega un agente con sus archivos y salida.
// Output queda siempre vacío hasta DASH-FEAT-013 (captura de output pendiente).
type AgentDetailDTO struct {
	Agent  AgentDTO  `json:"agent"`
	Files  []FileDTO `json:"files"`
	Output string    `json:"output"` // placeholder — captura real pendiente de DASH-FEAT-013
}

// FlowDTO representa el grafo dirigido del flujo de ejecución de un run.
type FlowDTO struct {
	Nodes []FlowNode `json:"nodes"`
	Edges []FlowEdge `json:"edges"`
}

// FlowNodeData contiene los datos de negocio del nodo expuestos al frontend.
type FlowNodeData struct {
	Label      string `json:"label"`
	Status     string `json:"status"`
	DurationMs *int64 `json:"durationMs"` // nil si el agente no terminó
	Tokens     *int   `json:"tokens"`     // nil si no hay datos
}

// FlowNode representa un nodo individual en el grafo de flujo de ejecución.
// La posición (x, y) la calcula el frontend con dagre; el backend envía ceros.
type FlowNode struct {
	ID   string       `json:"id"`
	Type string       `json:"type"` // siempre "agentNode" para nodos de agente
	Data FlowNodeData `json:"data"`
}

// FlowEdge representa una arista dirigida entre dos nodos del grafo de flujo.
type FlowEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// MetricsDTO contiene los agregados sobre los últimos 30 runs terminados.
type MetricsDTO struct {
	// RunsCount es el número de runs incluidos en el cálculo (0..30).
	// Solo cuenta runs con status IN ('success','failed'). Los 'running' se excluyen.
	RunsCount int `json:"runsCount"`

	// HasEnoughData es true cuando RunsCount >= 5.
	// El frontend usa este flag para decidir si renderiza los charts o el placeholder.
	HasEnoughData bool `json:"hasEnoughData"`

	// TotalTokens es la suma de runs.total_tokens del período actual.
	TotalTokens int `json:"totalTokens"`
	// TotalTokensDeltaPct es el cambio porcentual vs el período anterior.
	// nil cuando no hay período anterior con datos comparables.
	TotalTokensDeltaPct *float64 `json:"totalTokensDeltaPct"`

	// AvgDurationMs es el promedio de runs.duration_ms del período actual.
	AvgDurationMs int64 `json:"avgDurationMs"`
	// AvgDurationDeltaMs es la diferencia absoluta en ms vs período anterior.
	// Negativo = mejora (más rápido). nil cuando no hay comparación.
	AvgDurationDeltaMs *int64 `json:"avgDurationDeltaMs"`

	// SuccessRate está en el rango [0,1] (0.84 == 84%).
	SuccessRate float64 `json:"successRate"`
	// SuccessRateDelta es la diferencia absoluta en puntos porcentuales.
	// nil cuando no hay comparación.
	SuccessRateDelta *float64 `json:"successRateDelta"`

	// TokensPerRun son los runs del período actual ordenados cronológicamente ASC.
	// Nunca nil — siempre []TokensPerRunPoint{} cuando está vacío.
	TokensPerRun []TokensPerRunPoint `json:"tokensPerRun"`

	// AvgTimePerAgent agrupa por agent_role y promedia duration_ms.
	// Ordenado DESC por AvgDurationMs. Límite top 8.
	// Nunca nil — siempre []AgentTimePoint{} cuando está vacío.
	AvgTimePerAgent []AgentTimePoint `json:"avgTimePerAgent"`
}

// TokensPerRunPoint es una barra del chart "Tokens per run".
type TokensPerRunPoint struct {
	RunID     string `json:"runId"`
	StartedAt string `json:"startedAt"` // RFC3339Nano
	Tokens    int    `json:"tokens"`
}

// AgentTimePoint es una fila del chart "Avg time per agent".
type AgentTimePoint struct {
	Role          string `json:"role"`
	AvgDurationMs int64  `json:"avgDurationMs"`
	RunsIncluded  int    `json:"runsIncluded"`
}

// FileDTO representa un archivo producido o consumido por un agente.
type FileDTO struct {
	Path   string `json:"path"`
	Action string `json:"action"`
}
