package dto

// ToolUsageDTO is a count of tool invocations grouped by tool name.
type ToolUsageDTO struct {
	ToolName string `json:"toolName"`
	Count    int    `json:"count"`
}

// ToolUseDetailDTO is a single tool invocation with its input/command.
type ToolUseDetailDTO struct {
	ToolName  string `json:"toolName"`
	Command   string `json:"command"`
	AgentID   string `json:"agentId"`
	Timestamp string `json:"timestamp"`
}

// TurnStatsDTO contains activity statistics per turn.
type TurnStatsDTO struct {
	TurnNumber    int    `json:"turnNumber"`
	Prompt        string `json:"prompt"`
	Timestamp     string `json:"timestamp"`
	EndTimestamp  string `json:"endTimestamp"`
	FilesCount    int    `json:"filesCount"`
	ToolUsesCount int    `json:"toolUsesCount"`
	AgentsCount   int    `json:"agentsCount"`
}
