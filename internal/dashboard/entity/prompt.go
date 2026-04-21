package entity

// Prompt is a row from the prompts table.
type Prompt struct {
	Sequence  int
	Prompt    string
	Timestamp string
	Output    string
}
