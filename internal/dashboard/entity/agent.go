package entity

import "time"

// AgentRow is a row from the agents table prepared for flow view.
// Nullable fields use pointers to distinguish "zero" from "not set".
type AgentRow struct {
	AgentID     string
	AgentRole   string
	Sequence    *int
	DependsOn   string // JSON array serialized as TEXT in SQLite
	Status      string
	StartedAt   *time.Time
	EndedAt     *time.Time
	DurationMs  *int64
	TokensTotal *int
}

// AgentDetail is the full projection of an agent for the detail view.
type AgentDetail struct {
	AgentID      string
	AgentRole    string
	Status       string
	StartedAt    *time.Time
	EndedAt      *time.Time
	DurationMs   *int64
	TokensInput  *int
	TokensOutput *int
	TokensTotal  *int
	ExitCode     *int
	ErrorMsg     string
	Output       string
}
