package model

import "time"

type TaskStatus string

const (
	StatusTodo       TaskStatus = "[ ]"
	StatusInProgress TaskStatus = "[-]"
	StatusDone       TaskStatus = "[x]"
)

type Subtask struct {
	Status  TaskStatus
	Content string
}

// Task represents a single, raw task as it appears in the input markdown.
type Task struct {
	ID          string
	Title       string
	Project     string
	Status      TaskStatus
	StartDate   *time.Time
	EndDate     *time.Time
	Description []string
	SubTasks    []Subtask
}

type Section struct {
	Name  string
	Tasks []Task
}

// LogEntry represents a consolidated and formatted task
// for the final output. Multiple Task structs can be combined
// to create a single LogEntry.
type LogEntry struct {
	ID          string
	Title       string
	Project     string
	StartDate   *time.Time
	EndDate     *time.Time
	Description string
	SubTasks    []Subtask
}
