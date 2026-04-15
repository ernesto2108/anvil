package dto

// TaskDTO represents a task tracked within a run.
type TaskDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	CompletedAt string `json:"completedAt"`
}
