package entity

// FileRecord is a row from the files table.
type FileRecord struct {
	Path      string
	Operation string
	Diff      string
	AgentID   string // "" if edited directly by the main session
}
