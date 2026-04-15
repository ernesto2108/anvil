package dto

// ErrorsQuery contains filter params for listing error groups.
type ErrorsQuery struct {
	Status string `json:"status"`
	Search string `json:"search"`
}

// ErrorGroupDTO is the frontend representation of an error group.
type ErrorGroupDTO struct {
	ID               string `json:"id"`
	Fingerprint      string `json:"fingerprint"`
	Title            string `json:"title"`
	NormalizedMsg    string `json:"normalizedMsg"`
	ResolutionStatus string `json:"resolutionStatus"`
	FirstSeenAt      string `json:"firstSeenAt"`
	LastSeenAt       string `json:"lastSeenAt"`
	OccurrenceCount  int    `json:"occurrenceCount"`
	Notes            string `json:"notes"`
	CommitLink       string `json:"commitLink"`
}

// ErrorGroupDetailDTO is the full detail of an error group.
type ErrorGroupDetailDTO struct {
	Group   ErrorGroupDTO          `json:"group"`
	Runs    []ErrorGroupRunDTO     `json:"runs"`
	History []ErrorGroupHistoryDTO `json:"history"`
	Trend   []TrendPointDTO        `json:"trend"`
}

// ErrorGroupRunDTO represents a run linked to an error group.
type ErrorGroupRunDTO struct {
	RunID      string `json:"runId"`
	AgentName  string `json:"agentName"`
	ErrorMsg   string `json:"errorMsg"`
	ExitCode   *int   `json:"exitCode"`
	OccurredAt string `json:"occurredAt"`
}

// ErrorGroupHistoryDTO represents a resolution status change.
type ErrorGroupHistoryDTO struct {
	OldStatus  string `json:"oldStatus"`
	NewStatus  string `json:"newStatus"`
	Note       string `json:"note"`
	CommitLink string `json:"commitLink"`
	CreatedAt  string `json:"createdAt"`
}

// TrendPointDTO represents error occurrences on a single day.
type TrendPointDTO struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// ErrorsCountDTO contains summary counts of error groups.
type ErrorsCountDTO struct {
	New           int `json:"new"`
	Investigating int `json:"investigating"`
	WeekTotal     int `json:"weekTotal"`
}

// UpdateResolutionRequest is the request body for updating error resolution.
type UpdateResolutionRequest struct {
	GroupID    string `json:"groupId"`
	Status     string `json:"status"`
	Notes      string `json:"notes"`
	CommitLink string `json:"commitLink"`
}
