package dto

// FileDTO represents a file produced or consumed by an agent.
type FileDTO struct {
	Path    string `json:"path"`
	Action  string `json:"action"`
	Diff    string `json:"diff,omitempty"`
	AgentID string `json:"agentId"`
}
