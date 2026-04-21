package entity

// FileEdge represents a dependency edge between two files in a run.
// AgentID is "" if the edge was not attributed to a specific agent.
type FileEdge struct {
	SourcePath string
	TargetPath string
	Relation   string
	AgentID    string
}
