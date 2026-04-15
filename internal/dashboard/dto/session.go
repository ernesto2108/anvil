package dto

// SessionDetailDTO is the response for the session detail view.
type SessionDetailDTO struct {
	Run            RunDTO             `json:"run"`
	Files          []FileDTO          `json:"files"`
	Agents         []SessionAgentDTO  `json:"agents"`
	ActivityEvents []ActivityEventDTO `json:"activityEvents"`
	ToolUsage      []ToolUsageDTO     `json:"toolUsage"`
	ToolDetails    []ToolUseDetailDTO `json:"toolDetails"`
	Tasks          []TaskDTO          `json:"tasks"`
}
