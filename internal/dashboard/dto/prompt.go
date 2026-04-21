package dto

// PromptDTO is an individual prompt within a run.
type PromptDTO struct {
	Sequence  int    `json:"sequence"`
	Prompt    string `json:"prompt"`
	Timestamp string `json:"timestamp"`
	Output    string `json:"output"`
}
