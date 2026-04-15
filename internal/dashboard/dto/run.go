package dto

// RunsQuery contains filter parameters for listing runs.
type RunsQuery struct {
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
	Status    string `json:"status"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Project   string `json:"project"`
}

// RunDTO is the summarized representation of an individual run.
type RunDTO struct {
	ID               string `json:"id"`
	TaskID           string `json:"taskId"`
	TaskDesc         string `json:"taskDesc"`
	Status           string `json:"status"`
	Complexity       string `json:"complexity"`
	Provider         string `json:"provider"`
	Project          string `json:"project"`
	StartedAt        string `json:"startedAt"`
	EndedAt          string `json:"endedAt"`
	DurationMs       int64  `json:"durationMs"`
	FilesCount       int    `json:"filesCount"`
	AgentsCount      int    `json:"agentsCount"`
	ParentRunID      string `json:"parentRunId"`
	ChildrenCount    int    `json:"childrenCount"`
	Branch           string `json:"branch"`
	ErrorReason      string `json:"errorReason"`
	ToolUsesCount    int    `json:"toolUsesCount"`
	CompactionsCount int    `json:"compactionsCount"`
}

// RunDetailDTO combines a run with its agents.
type RunDetailDTO struct {
	Run    RunDTO     `json:"run"`
	Agents []AgentDTO `json:"agents"`
}
