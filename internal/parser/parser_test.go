package parser

import (
	"bufio"
	"strings"
	"testing"
	"time"

	"github.com/ahmaruff/tada/internal/model"
)

func TestParseContent(t *testing.T) {
	input := `## Backlog
- [ ] New task <!-- @crm -->
- [ ] New task 2 <!-- @hrm|#123 -->
- [x] New task 3 <!-- @hrm -->
- [x] Task done today <!-- @hrm|#124|2025-09-12 -->
- [x] Task done multi day <!-- @hrm|#124|2025-09-11 - 2025-09-12 -->

## Todo
### 2025-09-13 - Sabtu
- [ ] New task <!-- @crm -->
  i add some description
  - [-] also some subtask 
  - [-] also some subtask 2 
- [-] New task 2 <!-- @hrm|#123 -->
- [x] New task 3 <!-- @hrm -->

## Done
### 2025-09-12 - Jum'at
- [x] Task done today <!-- @hrm|#124 -->
- [x] Task done multi day <!-- @hrm|#124 -->

### 2025-09-11 - Kamis
- [-] Task done multi day <!-- @hrm|#124 -->

### 2025-09-10 - Rabu
- [x] old task <!-- @hrm|#121 -->
  i add description here too
- [x] old task 2 <!-- @crm -->
- [x] old task 3 <!-- @crm -->

## Archives
- [x] old task <!-- @hrm|#121|2025-09-09 - 2025-09-10 -->
  i add description here
  i add description here too
- [x] old task 2 <!-- @crm -->
- [x] some task`

	scanner := bufio.NewScanner(strings.NewReader(input))

	sections, err := ParseContent(scanner)

	if err != nil {
		t.Fatalf("ParseContent failed: %v", err)
	}

	// Test number of sections
	expectedSections := 4
	if len(sections) != expectedSections {
		t.Fatalf("Expected %d sections, got %d", expectedSections, len(sections))
	}

	// Test section names
	expectedNames := []string{"Backlog", "Todo", "Done", "Archives"}
	for i, section := range sections {
		if section.Name != expectedNames[i] {
			t.Errorf("Expected section %d name to be '%s', got '%s'", i, expectedNames[i], section.Name)
		}
	}

	// Test Backlog section
	backlogSection := sections[0]
	if len(backlogSection.Tasks) != 5 {
		t.Errorf("Expected 5 tasks in Backlog section, got %d", len(backlogSection.Tasks))
	}

	// Test first task in Backlog
	firstTask := backlogSection.Tasks[0]
	if firstTask.Title != "New task" {
		t.Errorf("Expected first task title to be 'New task', got '%s'", firstTask.Title)
	}

	if firstTask.Status != model.StatusTodo {
		t.Errorf("Expected first task status to be StatusTodo, got %v", firstTask.Status)
	}

	if firstTask.Project != "crm" {
		t.Errorf("Expected first task project to be 'crm', got '%s'", firstTask.Project)
	}

	// Test task with inline date
	taskWithDate := backlogSection.Tasks[3] // Task done today
	if taskWithDate.Project != "hrm" {
		t.Errorf("Expected task project to be 'hrm', got '%s'", taskWithDate.Project)
	}

	if taskWithDate.ID != "124" {
		t.Errorf("Expected task ID to be '124', got '%s'", taskWithDate.ID)
	}

	expectedDate := time.Date(2025, 9, 12, 0, 0, 0, 0, time.UTC)
	if taskWithDate.StartDate == nil || !taskWithDate.StartDate.Equal(expectedDate) {
		t.Errorf("Expected task start date to be %v, got %v", expectedDate, taskWithDate.StartDate)
	}

	if taskWithDate.EndDate == nil || !taskWithDate.EndDate.Equal(expectedDate) {
		t.Errorf("Expected task end date to be %v, got %v", expectedDate, taskWithDate.EndDate)
	}

	// Test task with date range
	taskWithRange := backlogSection.Tasks[4] // Task done multi day
	expectedStartDate := time.Date(2025, 9, 11, 0, 0, 0, 0, time.UTC)
	expectedEndDate := time.Date(2025, 9, 12, 0, 0, 0, 0, time.UTC)
	if taskWithRange.StartDate == nil || !taskWithRange.StartDate.Equal(expectedStartDate) {
		t.Errorf("Expected task start date to be %v, got %v", expectedStartDate, taskWithRange.StartDate)
	}

	if taskWithRange.EndDate == nil || !taskWithRange.EndDate.Equal(expectedEndDate) {
		t.Errorf("Expected task end date to be %v, got %v", expectedEndDate, taskWithRange.EndDate)
	}

	// Test Todo section with date headers
	todoSection := sections[1]
	if len(todoSection.Tasks) != 3 {
		t.Errorf("Expected 3 tasks in Todo section, got %d", len(todoSection.Tasks))
	}

	// Test task with header date fallback
	todoTask := todoSection.Tasks[0]
	expectedHeaderDate := time.Date(2025, 9, 13, 0, 0, 0, 0, time.UTC)
	if todoTask.StartDate == nil || !todoTask.StartDate.Equal(expectedHeaderDate) {
		t.Errorf("Expected todo task start date to be %v, got %v", expectedHeaderDate, todoTask.StartDate)
	}

	// Test task with description and subtasks
	taskWithSubs := todoSection.Tasks[0]
	if len(taskWithSubs.Description) != 1 {
		t.Errorf("Expected 1 description line, got %d", len(taskWithSubs.Description))
	}

	if taskWithSubs.Description[0] != "i add some description" {
		t.Errorf("Expected description to be 'i add some description', got '%s'", taskWithSubs.Description[0])
	}

	if len(taskWithSubs.SubTasks) != 2 {
		t.Errorf("Expected 2 subtasks, got %d", len(taskWithSubs.SubTasks))
	}

	if taskWithSubs.SubTasks[0].Status != model.StatusInProgress {
		t.Errorf("Expected subtask status to be StatusInProgress, got %v", taskWithSubs.SubTasks[0].Status)
	}

	if taskWithSubs.SubTasks[0].Content != "also some subtask" {
		t.Errorf("Expected subtask content to be 'also some subtask', got '%s'", taskWithSubs.SubTasks[0].Content)
	}

	// Test Done section with multiple date groups
	doneSection := sections[2]
	if len(doneSection.Tasks) != 6 {
		t.Errorf("Expected 5 tasks in Done section, got %d", len(doneSection.Tasks))
	}

	// Test Archives section
	archivesSection := sections[3]
	if len(archivesSection.Tasks) != 3 {
		t.Errorf("Expected 3 tasks in Archives section, got %d", len(archivesSection.Tasks))
	}

	// Test task with multi-line description
	taskWithMultiDesc := archivesSection.Tasks[0]
	if len(taskWithMultiDesc.Description) != 2 {
		t.Errorf("Expected 2 description lines, got %d", len(taskWithMultiDesc.Description))
	}
	if taskWithMultiDesc.Description[0] != "i add description here" {
		t.Errorf("Expected first description line to be 'i add description here', got '%s'", taskWithMultiDesc.Description[0])
	}
	if taskWithMultiDesc.Description[1] != "i add description here too" {
		t.Errorf("Expected second description line to be 'i add description here too', got '%s'", taskWithMultiDesc.Description[1])
	}

	// Test task without comment
	taskWithoutComment := archivesSection.Tasks[2]
	if taskWithoutComment.Title != "some task" {
		t.Errorf("Expected task title to be 'some task', got '%s'", taskWithoutComment.Title)
	}
	if taskWithoutComment.Project != "" {
		t.Errorf("Expected task project to be empty, got '%s'", taskWithoutComment.Project)
	}
	if taskWithoutComment.ID != "" {
		t.Errorf("Expected task ID to be empty, got '%s'", taskWithoutComment.ID)
	}
}

func TestParseComment(t *testing.T) {
	tests := []struct {
		name      string
		comment   string
		project   string
		taskId    string
		startDate *time.Time
		endDate   *time.Time
	}{
		{
			name:    "empty comment",
			comment: "",
			project: "",
			taskId:  "",
		},
		{
			name:    "project only",
			comment: "@crm",
			project: "crm",
			taskId:  "",
		},
		{
			name:    "project and id",
			comment: "@hrm|#123",
			project: "hrm",
			taskId:  "123",
		},
		{
			name:      "project, id, and single date",
			comment:   "@hrm|#124|2025-09-12",
			project:   "hrm",
			taskId:    "124",
			startDate: timePtr(2025, 9, 12),
			endDate:   timePtr(2025, 9, 12),
		},
		{
			name:      "project, id, and date range",
			comment:   "@hrm|#124|2025-09-11 - 2025-09-12",
			project:   "hrm",
			taskId:    "124",
			startDate: timePtr(2025, 9, 11),
			endDate:   timePtr(2025, 9, 12),
		},
		{
			name:      "date range only",
			comment:   "2025-09-09 - 2025-09-10",
			project:   "",
			taskId:    "",
			startDate: timePtr(2025, 9, 9),
			endDate:   timePtr(2025, 9, 10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, taskId, startDate, endDate := parseComment(tt.comment)

			if project != tt.project {
				t.Errorf("Expected project '%s', got '%s'", tt.project, project)
			}

			if taskId != tt.taskId {
				t.Errorf("Expected taskId '%s', got '%s'", tt.taskId, taskId)
			}

			if tt.startDate == nil && startDate != nil {
				t.Errorf("Expected startDate to be nil, got %v", startDate)
			}
			if tt.startDate != nil && (startDate == nil || !startDate.Equal(*tt.startDate)) {
				t.Errorf("Expected startDate %v, got %v", tt.startDate, startDate)
			}

			if tt.endDate == nil && endDate != nil {
				t.Errorf("Expected endDate to be nil, got %v", endDate)
			}
			if tt.endDate != nil && (endDate == nil || !endDate.Equal(*tt.endDate)) {
				t.Errorf("Expected endDate %v, got %v", tt.endDate, endDate)
			}
		})
	}
}

func TestCheckLineType(t *testing.T) {
	tests := []struct {
		line          string
		expectedType  LineType
		expectedValue string
	}{
		{"## Backlog", LineSectionHeader, "Backlog"},
		{"##   Todo   ", LineSectionHeader, "Todo"},
		{"### 2025-09-13 - Sabtu", LineDateHeader, "2025-09-13"},
		{"### 2025-09-12", LineDateHeader, "2025-09-12"},
		{"- [ ] New task", LineTask, ""},
		{"- [x] Done task <!-- @crm -->", LineTask, ""},
		{"- [-] In progress task", LineTask, ""},
		{"  - [ ] Subtask", LineSubtask, ""},
		{"  - [x] Done subtask", LineSubtask, ""},
		{"  some description", LineDescription, "some description"},
		{"    more description", LineDescription, "more description"},
		{"", LineUnknown, ""},
		{"random text", LineUnknown, ""},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			lineType, value := checkLineType(tt.line)
			if lineType != tt.expectedType {
				t.Errorf("Expected line type %v, got %v", tt.expectedType, lineType)
			}
			if value != tt.expectedValue {
				t.Errorf("Expected value '%s', got '%s'", tt.expectedValue, value)
			}
		})
	}
}

func TestParseTaskLine(t *testing.T) {
	fallbackDate := timePtr(2025, 9, 13)

	tests := []struct {
		name         string
		line         string
		fallbackDate *time.Time
		expectedTask model.Task
	}{
		{
			name: "basic task with project",
			line: "- [ ] New task <!-- @crm -->",
			expectedTask: model.Task{
				Title:   "New task",
				Project: "crm",
				Status:  model.StatusTodo,
			},
		},
		{
			name:         "task with fallback date",
			line:         "- [x] Done task <!-- @hrm -->",
			fallbackDate: fallbackDate,
			expectedTask: model.Task{
				Title:     "Done task",
				Project:   "hrm",
				Status:    model.StatusDone,
				StartDate: fallbackDate,
				EndDate:   fallbackDate,
			},
		},
		{
			name: "task with inline date",
			line: "- [-] Task with date <!-- @hrm|#123|2025-09-12 -->",
			expectedTask: model.Task{
				Title:     "Task with date",
				Project:   "hrm",
				ID:        "123",
				Status:    model.StatusInProgress,
				StartDate: timePtr(2025, 9, 12),
				EndDate:   timePtr(2025, 9, 12),
			},
		},
		{
			name: "task without comment",
			line: "- [x] Simple task",
			expectedTask: model.Task{
				Title:  "Simple task",
				Status: model.StatusDone,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := parseTaskLine(tt.line, tt.fallbackDate)

			if task.Title != tt.expectedTask.Title {
				t.Errorf("Expected title '%s', got '%s'", tt.expectedTask.Title, task.Title)
			}
			if task.Project != tt.expectedTask.Project {
				t.Errorf("Expected project '%s', got '%s'", tt.expectedTask.Project, task.Project)
			}
			if task.ID != tt.expectedTask.ID {
				t.Errorf("Expected ID '%s', got '%s'", tt.expectedTask.ID, task.ID)
			}
			if task.Status != tt.expectedTask.Status {
				t.Errorf("Expected status %v, got %v", tt.expectedTask.Status, task.Status)
			}

			if !timePtrEqual(task.StartDate, tt.expectedTask.StartDate) {
				t.Errorf("Expected start date %v, got %v", tt.expectedTask.StartDate, task.StartDate)
			}
			if !timePtrEqual(task.EndDate, tt.expectedTask.EndDate) {
				t.Errorf("Expected end date %v, got %v", tt.expectedTask.EndDate, task.EndDate)
			}
		})
	}
}

func TestParseSubTaskLine(t *testing.T) {
	tests := []struct {
		line     string
		expected model.Subtask
	}{
		{
			line: "  - [ ] Some subtask",
			expected: model.Subtask{
				Status:  model.StatusTodo,
				Content: "Some subtask",
			},
		},
		{
			line: "    - [x] Done subtask",
			expected: model.Subtask{
				Status:  model.StatusDone,
				Content: "Done subtask",
			},
		},
		{
			line: "  - [-] In progress subtask",
			expected: model.Subtask{
				Status:  model.StatusInProgress,
				Content: "In progress subtask",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			subtask := parseSubTaskLine(tt.line)

			if subtask.Status != tt.expected.Status {
				t.Errorf("Expected status %v, got %v", tt.expected.Status, subtask.Status)
			}
			if subtask.Content != tt.expected.Content {
				t.Errorf("Expected content '%s', got '%s'", tt.expected.Content, subtask.Content)
			}
		})
	}
}

// Helper functions
func timePtr(year, month, day int) *time.Time {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return &t
}

func timePtrEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}
