package dto

// AgentDTO is the full representation of an agent within a run.
type AgentDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	DurationMs *int64 `json:"durationMs"`
	StartedAt  string `json:"startedAt"`
	EndedAt    string `json:"endedAt"`
	ErrorMsg   string `json:"errorMsg"`
}

// AgentDetailDTO combines an agent with its files and output.
type AgentDetailDTO struct {
	Agent  AgentDTO  `json:"agent"`
	Files  []FileDTO `json:"files"`
	Output string    `json:"output"`
}

// SessionAgentDTO is a lightweight agent summary for the session detail view.
type SessionAgentDTO struct {
	ID         string `json:"id"`
	Role       string `json:"role"`
	Status     string `json:"status"`
	StartedAt  string `json:"startedAt"`
	EndedAt    string `json:"endedAt"`
	DurationMs *int64 `json:"durationMs"`
	Output     string `json:"output"`
}
