package processor

import (
	"testing"
	"time"

	"github.com/ahmaruff/tada/internal/model"
)

func TestConsolidateTasks(t *testing.T) {
	sections := []model.Section{
		{
			Name: model.SectionBacklog,
			Tasks: []model.Task{
				{
					ID:      "123",
					Title:   "Task with updates",
					Project: "hrm",
					Status:  model.StatusTodo,
				},
				{
					ID:      "124",
					Title:   "Multi-day task",
					Project: "hrm",
					Status:  model.StatusTodo,
				},
				{
					ID:      "125",
					Title:   "Task without updates",
					Project: "crm",
					Status:  model.StatusTodo,
				},
				{
					Title:   "Task without ID",
					Project: "crm",
					Status:  model.StatusTodo,
				},
			},
		},
		{
			Name: model.SectionTodo,
			Tasks: []model.Task{
				{
					ID:          "123",
					Title:       "Task with updates",
					Project:     "hrm",
					Status:      model.StatusInProgress,
					StartDate:   timePtr(2025, 9, 13),
					EndDate:     timePtr(2025, 9, 13),
					Description: []string{"Added description in todo"},
					SubTasks: []model.Subtask{
						{Status: model.StatusTodo, Content: "Subtask 1"},
					},
				},
			},
		},
		{
			Name: model.SectionDone,
			Tasks: []model.Task{
				{
					ID:          "123",
					Title:       "Task with updates",
					Project:     "hrm",
					Status:      model.StatusDone,
					StartDate:   timePtr(2025, 9, 12), // Earlier start date
					EndDate:     timePtr(2025, 9, 14), // Later end date
					Description: []string{"Added description in done", "Another description"},
					SubTasks: []model.Subtask{
						{Status: model.StatusDone, Content: "Subtask 1"}, // Same content, different status
						{Status: model.StatusTodo, Content: "Subtask 2"}, // New subtask
					},
				},
				{
					ID:        "124",
					Title:     "Multi-day task",
					Project:   "hrm",
					Status:    model.StatusDone,
					StartDate: timePtr(2025, 9, 11),
					EndDate:   timePtr(2025, 9, 12),
				},
			},
		},
		{
			Name: model.SectionArchives,
			Tasks: []model.Task{
				{
					ID:      "999",
					Title:   "Old task",
					Project: "old",
					Status:  model.StatusDone,
				},
			},
		},
	}

	result := ConsolidateTasks(sections)

	// Test that we still have 4 sections
	if len(result) != 4 {
		t.Fatalf("Expected 4 sections, got %d", len(result))
	}

	// Test Backlog section updates
	backlogSection := result[0]
	if backlogSection.Name != model.SectionBacklog {
		t.Errorf("Expected first section to be Backlog, got %s", backlogSection.Name)
	}

	// Test task 123 - should be updated with merged data
	task123 := backlogSection.Tasks[0]
	if task123.ID != "123" {
		t.Errorf("Expected task ID to be 123, got %s", task123.ID)
	}

	if task123.Status != model.StatusDone {
		t.Errorf("Expected task status to be StatusDone, got %v", task123.Status)
	}

	// Check merged dates - should have earliest start and latest end
	expectedStart := timePtr(2025, 9, 12)
	expectedEnd := timePtr(2025, 9, 14)
	if !timePtrEqual(task123.StartDate, expectedStart) {
		t.Errorf("Expected start date %v, got %v", expectedStart, task123.StartDate)
	}

	if !timePtrEqual(task123.EndDate, expectedEnd) {
		t.Errorf("Expected end date %v, got %v", expectedEnd, task123.EndDate)
	}

	// Check merged descriptions
	expectedDescriptions := []string{"Added description in todo", "Added description in done", "Another description"}
	if len(task123.Description) != len(expectedDescriptions) {
		t.Errorf("Expected %d descriptions, got %d", len(expectedDescriptions), len(task123.Description))
	}

	for i, desc := range expectedDescriptions {
		if i < len(task123.Description) && task123.Description[i] != desc {
			t.Errorf("Expected description[%d] to be '%s', got '%s'", i, desc, task123.Description[i])
		}
	}

	// Check merged subtasks
	if len(task123.SubTasks) != 2 {
		t.Errorf("Expected 2 subtasks, got %d", len(task123.SubTasks))
	}

	// First subtask should have status updated to Done
	if task123.SubTasks[0].Status != model.StatusDone {
		t.Errorf("Expected first subtask status to be StatusDone, got %v", task123.SubTasks[0].Status)
	}
	if task123.SubTasks[0].Content != "Subtask 1" {
		t.Errorf("Expected first subtask content to be 'Subtask 1', got '%s'", task123.SubTasks[0].Content)
	}

	// Test task 124 - should be updated
	task124 := backlogSection.Tasks[1]
	if task124.Status != model.StatusDone {
		t.Errorf("Expected task 124 status to be StatusDone, got %v", task124.Status)
	}

	if !timePtrEqual(task124.StartDate, timePtr(2025, 9, 11)) {
		t.Errorf("Expected task 124 start date to be updated")
	}

	// Test task 125 - should remain unchanged (no updates found)
	task125 := backlogSection.Tasks[2]
	if task125.Status != model.StatusTodo {
		t.Errorf("Expected task 125 status to remain StatusTodo, got %v", task125.Status)
	}

	if task125.StartDate != nil {
		t.Errorf("Expected task 125 start date to remain nil, got %v", task125.StartDate)
	}

	// Test task without ID - should remain unchanged
	taskNoID := backlogSection.Tasks[3]
	if taskNoID.Status != model.StatusTodo {
		t.Errorf("Expected task without ID to remain StatusTodo, got %v", taskNoID.Status)
	}

	// Test that Todo, Done, Archives sections remain unchanged
	todoSection := result[1]
	if len(todoSection.Tasks) != 1 {
		t.Errorf("Expected Todo section to have 1 task, got %d", len(todoSection.Tasks))
	}

	doneSection := result[2]
	if len(doneSection.Tasks) != 2 {
		t.Errorf("Expected Done section to have 2 tasks, got %d", len(doneSection.Tasks))
	}

	archivesSection := result[3]
	if len(archivesSection.Tasks) != 1 {
		t.Errorf("Expected Archives section to have 1 task, got %d", len(archivesSection.Tasks))
	}
}

func TestMergeTaskData(t *testing.T) {
	existing := &model.Task{
		ID:          "123",
		Title:       "Test Task",
		Status:      model.StatusTodo,
		StartDate:   timePtr(2025, 9, 13),
		EndDate:     timePtr(2025, 9, 13),
		Description: []string{"Original description"},
		SubTasks: []model.Subtask{
			{Status: model.StatusTodo, Content: "Original subtask"},
		},
	}

	new := &model.Task{
		ID:          "123",
		Title:       "Test Task",
		Status:      model.StatusDone,
		StartDate:   timePtr(2025, 9, 12), // Earlier
		EndDate:     timePtr(2025, 9, 14), // Later
		Description: []string{"New description"},
		SubTasks: []model.Subtask{
			{Status: model.StatusDone, Content: "Original subtask"}, // Same content, different status
			{Status: model.StatusTodo, Content: "New subtask"},      // New subtask
		},
	}

	result := mergeTaskData(existing, new)

	// Status should be updated to Done
	if result.Status != model.StatusDone {
		t.Errorf("Expected status StatusDone, got %v", result.Status)
	}

	// Start date should be earliest
	expectedStart := timePtr(2025, 9, 12)
	if !timePtrEqual(result.StartDate, expectedStart) {
		t.Errorf("Expected start date %v, got %v", expectedStart, result.StartDate)
	}

	// End date should be latest
	expectedEnd := timePtr(2025, 9, 14)
	if !timePtrEqual(result.EndDate, expectedEnd) {
		t.Errorf("Expected end date %v, got %v", expectedEnd, result.EndDate)
	}

	// Should have both descriptions
	if len(result.Description) != 2 {
		t.Errorf("Expected 2 descriptions, got %d", len(result.Description))
	}

	// Should have both subtasks, with first one updated to Done
	if len(result.SubTasks) != 2 {
		t.Errorf("Expected 2 subtasks, got %d", len(result.SubTasks))
	}
	if result.SubTasks[0].Status != model.StatusDone {
		t.Errorf("Expected first subtask to be Done, got %v", result.SubTasks[0].Status)
	}
}

func TestMergeDescriptions(t *testing.T) {
	tests := []struct {
		name     string
		existing []string
		new      []string
		expected []string
	}{
		{
			name:     "both empty",
			existing: []string{},
			new:      []string{},
			expected: []string{},
		},
		{
			name:     "existing empty",
			existing: []string{},
			new:      []string{"new desc"},
			expected: []string{"new desc"},
		},
		{
			name:     "new empty",
			existing: []string{"existing desc"},
			new:      []string{},
			expected: []string{"existing desc"},
		},
		{
			name:     "no duplicates",
			existing: []string{"desc1"},
			new:      []string{"desc2"},
			expected: []string{"desc1", "desc2"},
		},
		{
			name:     "with duplicates",
			existing: []string{"desc1", "desc2"},
			new:      []string{"desc2", "desc3"},
			expected: []string{"desc1", "desc2", "desc3"},
		},
		{
			name:     "empty strings filtered",
			existing: []string{"desc1", ""},
			new:      []string{"", "desc2"},
			expected: []string{"desc1", "desc2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeDescriptions(tt.existing, tt.new)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d descriptions, got %d", len(tt.expected), len(result))
			}
			for i, expected := range tt.expected {
				if i < len(result) && result[i] != expected {
					t.Errorf("Expected description[%d] to be '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}
}

func TestMergeSubtasks(t *testing.T) {
	existing := []model.Subtask{
		{Status: model.StatusTodo, Content: "subtask1"},
		{Status: model.StatusInProgress, Content: "subtask2"},
	}

	new := []model.Subtask{
		{Status: model.StatusDone, Content: "subtask1"}, // Same content, higher priority status
		{Status: model.StatusTodo, Content: "subtask3"}, // New subtask
		{Status: model.StatusTodo, Content: "subtask2"}, // Same content, lower priority status
	}

	result := mergeSubtasks(existing, new)

	if len(result) != 3 {
		t.Errorf("Expected 3 subtasks, got %d", len(result))
	}

	// Find subtask1 - should have status updated to Done
	var subtask1 *model.Subtask
	for i := range result {
		if result[i].Content == "subtask1" {
			subtask1 = &result[i]
			break
		}
	}
	if subtask1 == nil {
		t.Error("subtask1 not found in result")
	} else if subtask1.Status != model.StatusDone {
		t.Errorf("Expected subtask1 status to be Done, got %v", subtask1.Status)
	}

	// Find subtask2 - should keep InProgress status (higher than Todo)
	var subtask2 *model.Subtask
	for i := range result {
		if result[i].Content == "subtask2" {
			subtask2 = &result[i]
			break
		}
	}
	if subtask2 == nil {
		t.Error("subtask2 not found in result")
	} else if subtask2.Status != model.StatusInProgress {
		t.Errorf("Expected subtask2 status to remain InProgress, got %v", subtask2.Status)
	}
}

// Helper function
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
