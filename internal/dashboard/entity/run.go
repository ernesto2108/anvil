package entity

import "time"

// RunSummary is a domain representation of a run row.
// Nullable fields use pointers to distinguish "zero" from "not set".
type RunSummary struct {
	ID            string
	TaskID        string
	TaskDesc      string
	Status        string
	Complexity    string
	Provider      string
	Project       string
	StartedAt     time.Time
	EndedAt       *time.Time
	DurationMs    *int64
	TotalTokens   int
	FilesCount    int
	AgentsCount   int
	QAScore       *float64
	ParentRunID   string
	ChildrenCount int
	Branch        string
	ErrorReason   string
}
