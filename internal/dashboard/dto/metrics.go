package dto

// MetricsDTO contains aggregated metrics for the dashboard overview.
type MetricsDTO struct {
	RunsCount     int  `json:"runsCount"`
	HasEnoughData bool `json:"hasEnoughData"`

	TotalTokens   int     `json:"totalTokens"`
	AvgDurationMs int64   `json:"avgDurationMs"`
	SuccessRate   float64 `json:"successRate"`

	TotalTokensDeltaPct *float64 `json:"totalTokensDeltaPct"`
	AvgDurationDeltaMs  *int64   `json:"avgDurationDeltaMs"`
	SuccessRateDelta    *float64 `json:"successRateDelta"`

	TokensPerRun    []TokensPerRunDTO `json:"tokensPerRun"`
	AvgTimePerAgent []AgentTimeDTO    `json:"avgTimePerAgent"`
}

// TokensPerRunDTO is a data point for the "Tokens per run" chart.
type TokensPerRunDTO struct {
	RunID     string `json:"runId"`
	StartedAt string `json:"startedAt"`
	Tokens    int    `json:"tokens"`
}

// AgentTimeDTO is a data point for the "Avg time per agent" chart.
type AgentTimeDTO struct {
	Role          string `json:"role"`
	AvgDurationMs int64  `json:"avgDurationMs"`
	RunsIncluded  int    `json:"runsIncluded"`
}
