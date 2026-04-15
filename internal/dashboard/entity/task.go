package entity

// Task is a row from the tasks table.
type Task struct {
	TaskID      string
	Title       string
	Status      string
	CreatedAt   string
	CompletedAt string
}
