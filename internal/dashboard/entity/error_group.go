package entity

import "time"

// ErrorGroup is the domain representation of an error group.
type ErrorGroup struct {
	ID               string
	Fingerprint      string
	Title            string
	NormalizedMsg    string
	ResolutionStatus string
	FirstSeenAt      time.Time
	LastSeenAt       time.Time
	OccurrenceCount  int
	Notes            string
	CommitLink       string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ErrorGroupRun represents a run linked to an error group.
type ErrorGroupRun struct {
	RunID      string
	AgentName  string
	ErrorMsg   string
	ExitCode   *int
	OccurredAt string
}

// ErrorGroupHistory represents a resolution status change record.
type ErrorGroupHistory struct {
	ID         int
	OldStatus  string
	NewStatus  string
	Note       string
	CommitLink string
	CreatedAt  string
}

// TrendPoint represents error occurrences on a single day.
type TrendPoint struct {
	Date  string // YYYY-MM-DD
	Count int
}

// ErrorCounts contains summary counts of error groups by status.
type ErrorCounts struct {
	New           int
	Investigating int
	WeekTotal     int
}
