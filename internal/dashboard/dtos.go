//go:build dashboard

package dashboard

// DTOs expuestos al frontend via Wails bindings.

// RunsQuery contiene los parámetros de filtro para listar runs.
type RunsQuery struct {
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
	Status    string `json:"status"`
	StartDate string `json:"startDate"` // RFC3339, opcional
	EndDate   string `json:"endDate"`   // RFC3339, opcional
	Project   string `json:"project"`   // opcional
}

// RunDTO es la representación resumida de un run individual.
type RunDTO struct {
	ID            string `json:"id"`
	TaskID        string `json:"taskId"`
	TaskDesc      string `json:"taskDesc"`
	Status        string `json:"status"`
	Complexity    string `json:"complexity"`
	Provider      string `json:"provider"`
	Project       string `json:"project"`
	StartedAt     string `json:"startedAt"`   // RFC3339Nano
	EndedAt       string `json:"endedAt"`     // "" si todavía no terminó
	DurationMs    int64  `json:"durationMs"`  // 0 si todavía no terminó
	FilesCount    int    `json:"filesCount"`
	AgentsCount   int    `json:"agentsCount"`
	ParentRunID      string `json:"parentRunId"`
	ChildrenCount    int    `json:"childrenCount"`
	Branch           string `json:"branch"`
	ErrorReason      string `json:"errorReason"`
	ToolUsesCount    int    `json:"toolUsesCount"`
	CompactionsCount int    `json:"compactionsCount"`
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
	DurationMs *int64 `json:"durationMs"` // nil si el agente aún está corriendo
	StartedAt  string `json:"startedAt"` // RFC3339Nano, "" si no seteado
	EndedAt     string `json:"endedAt"`     // RFC3339Nano, "" si no terminó
	ErrorMsg    string `json:"errorMsg"`    // "" si no falló
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

// ToolUsageDTO is a count of tool invocations grouped by tool name.
type ToolUsageDTO struct {
	ToolName string `json:"toolName"`
	Count    int    `json:"count"`
}

// ToolUseDetailDTO is a single tool invocation with its input/command.
type ToolUseDetailDTO struct {
	ToolName  string `json:"toolName"`
	Command   string `json:"command"`   // Extracted command (for Bash) or summary
	AgentID   string `json:"agentId"`
	Timestamp string `json:"timestamp"` // RFC3339Nano
}

// TaskDTO represents a task tracked within a run.
type TaskDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Status      string `json:"status"` // pending | completed
	CreatedAt   string `json:"createdAt"`
	CompletedAt string `json:"completedAt"`
}

// SessionDetailDTO is the response for the session detail view.
// Combines run summary, files changed, and agent summaries.
type SessionDetailDTO struct {
	Run            RunDTO              `json:"run"`
	Files          []FileDTO           `json:"files"`
	Agents         []SessionAgentDTO   `json:"agents"`
	ActivityEvents []ActivityEventDTO  `json:"activityEvents"`
	ToolUsage      []ToolUsageDTO      `json:"toolUsage"`
	ToolDetails    []ToolUseDetailDTO  `json:"toolDetails"`
	Tasks          []TaskDTO           `json:"tasks"`
}

// ActivityEventDTO is a file edit event with timestamp and agent attribution.
type ActivityEventDTO struct {
	Timestamp string `json:"timestamp"` // RFC3339Nano
	AgentID   string `json:"agentId"`   // "" = main session (Claude direct)
}

// SessionAgentDTO is a lightweight agent summary for the session detail view.
type SessionAgentDTO struct {
	ID         string `json:"id"`
	Role       string `json:"role"`
	Status     string `json:"status"`
	StartedAt  string `json:"startedAt"`  // RFC3339Nano, "" si no seteado
	EndedAt    string `json:"endedAt"`    // RFC3339Nano, "" si no terminó
	DurationMs *int64 `json:"durationMs"`
	Output     string `json:"output"`
}

// FileDTO representa un archivo producido o consumido por un agente.
type FileDTO struct {
	Path    string `json:"path"`
	Action  string `json:"action"`
	Diff    string `json:"diff,omitempty"`
	AgentID string `json:"agentId"` // "" = editado por sesión principal (Claude directo)
}

// PromptDTO es un prompt individual dentro de un run.
type PromptDTO struct {
	Sequence  int    `json:"sequence"`
	Prompt    string `json:"prompt"`
	Timestamp string `json:"timestamp"` // RFC3339Nano
}

// TurnStatsDTO son las estadísticas de actividad por turn.
type TurnStatsDTO struct {
	TurnNumber    int    `json:"turnNumber"`
	Prompt        string `json:"prompt"`
	Timestamp     string `json:"timestamp"`    // inicio del turn
	EndTimestamp  string `json:"endTimestamp"` // fin del turn
	FilesCount    int    `json:"filesCount"`
	ToolUsesCount int    `json:"toolUsesCount"`
	AgentsCount   int    `json:"agentsCount"`
}
