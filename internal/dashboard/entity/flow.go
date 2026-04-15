package entity

// FlowGraph represents the directed graph of a run's execution flow.
type FlowGraph struct {
	Nodes []FlowNode
	Edges []FlowEdge
}

// FlowNode represents an individual node in the execution flow graph.
type FlowNode struct {
	ID         string
	Label      string
	Status     string
	DurationMs *int64
}

// FlowEdge represents a directed edge between two nodes in the flow graph.
type FlowEdge struct {
	ID     string
	Source string
	Target string
}
