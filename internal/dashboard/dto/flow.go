package dto

// FlowDTO represents the directed graph of a run's execution flow.
type FlowDTO struct {
	Nodes []FlowNodeDTO `json:"nodes"`
	Edges []FlowEdgeDTO `json:"edges"`
}

// FlowNodeDataDTO contains the business data of a node exposed to the frontend.
type FlowNodeDataDTO struct {
	Label      string `json:"label"`
	Status     string `json:"status"`
	DurationMs *int64 `json:"durationMs"`
}

// FlowNodeDTO represents an individual node in the execution flow graph.
type FlowNodeDTO struct {
	ID   string           `json:"id"`
	Type string           `json:"type"`
	Data FlowNodeDataDTO  `json:"data"`
}

// FlowEdgeDTO represents a directed edge between two flow graph nodes.
type FlowEdgeDTO struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}
