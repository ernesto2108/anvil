package entity

// ToolUsage is a count of tool invocations grouped by tool name.
type ToolUsage struct {
	ToolName string
	Count    int
}

// ToolUseDetail is a single tool invocation with its input.
type ToolUseDetail struct {
	ToolName  string
	ToolInput string
	AgentID   string
	Timestamp string
}

// TurnStats contains activity statistics for a turn (period between consecutive prompts).
type TurnStats struct {
	TurnNumber    int
	Prompt        string
	Timestamp     string // start of the turn (= prompt timestamp)
	EndTimestamp  string // start of the next turn, or run ended_at
	FilesCount    int
	ToolUsesCount int
	AgentsCount   int
}
