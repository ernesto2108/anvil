package dto

// ActivityEventDTO is a file edit event with timestamp and agent attribution.
type ActivityEventDTO struct {
	Timestamp string `json:"timestamp"`
	AgentID   string `json:"agentId"`
	Path      string `json:"path"`
}
