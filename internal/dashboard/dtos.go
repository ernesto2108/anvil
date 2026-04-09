//go:build dashboard

package dashboard

// DTOs expuestos al frontend via Wails bindings. Son stubs minimos — la
// definicion completa se agrega en DASH-FEAT-005+ conforme se implementen las vistas.

// RunsQuery holds filter parameters for listing runs.
type RunsQuery struct {
	Limit  int    `json:"limit"`
	Status string `json:"status"`
}

// RunDTO is the summary representation of a single run.
type RunDTO struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// RunDetailDTO aggregates a run with its agents.
type RunDetailDTO struct {
	Run    RunDTO     `json:"run"`
	Agents []AgentDTO `json:"agents"`
}

// AgentDTO is the summary representation of an agent within a run.
type AgentDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AgentDetailDTO aggregates an agent with its files and output.
type AgentDetailDTO struct {
	Agent  AgentDTO  `json:"agent"`
	Files  []FileDTO `json:"files"`
	Output string    `json:"output"`
}

// FlowDTO represents the directed graph of a run's execution flow.
type FlowDTO struct {
	Nodes []FlowNode `json:"nodes"`
	Edges []FlowEdge `json:"edges"`
}

// FlowNode represents a single node in the execution flow graph.
type FlowNode struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Status string `json:"status"`
}

// FlowEdge represents a directed edge between two nodes in the flow graph.
type FlowEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// MetricsQuery holds the time range for a metrics request.
type MetricsQuery struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// MetricsDTO holds aggregated run metrics for a given time range.
type MetricsDTO struct {
	TotalRuns   int     `json:"totalRuns"`
	SuccessRate float64 `json:"successRate"`
}

// FileDTO represents a file produced or consumed by an agent.
type FileDTO struct {
	Path   string `json:"path"`
	Action string `json:"action"`
}
