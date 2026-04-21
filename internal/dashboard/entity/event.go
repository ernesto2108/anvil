package entity

// ActivityEvent is a timestamp + agent attribution for a file edit event.
type ActivityEvent struct {
	Timestamp string
	AgentID   string // "" = main session (Claude direct)
	Path      string
}
