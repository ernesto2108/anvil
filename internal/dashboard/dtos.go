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

// AgentDTO es la representación resumida de un agente dentro de un run.
type AgentDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AgentDetailDTO agrega un agente con sus archivos y salida.
type AgentDetailDTO struct {
	Agent  AgentDTO  `json:"agent"`
	Files  []FileDTO `json:"files"`
	Output string    `json:"output"`
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

// MetricsQuery contiene el rango de tiempo para una petición de métricas.
type MetricsQuery struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// MetricsDTO contiene métricas agregadas de runs para un rango de tiempo dado.
type MetricsDTO struct {
	TotalRuns   int     `json:"totalRuns"`
	SuccessRate float64 `json:"successRate"`
}

// FileDTO representa un archivo producido o consumido por un agente.
type FileDTO struct {
	Path   string `json:"path"`
	Action string `json:"action"`
}
