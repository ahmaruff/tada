package model

import "time"

type TaskStatus string

const (
	StatusTodo       TaskStatus = "[ ]"
	StatusInProgress TaskStatus = "[-]"
	StatusDone       TaskStatus = "[x]"
)

type SectionName string

const (
	SectionBacklog  SectionName = "Backlog"
	SectionArchives SectionName = "Archives"
	SectionTodo     SectionName = "Todo"
	SectionDone     SectionName = "Done"
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
	Name  SectionName
	Tasks []Task
}
