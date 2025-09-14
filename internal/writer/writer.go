package writer

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ahmaruff/tada/internal/model"
)

func WriteInputFile(sections []model.Section, filePath string) error {
	content := GenerateInputMarkdown(sections)
	return os.WriteFile(filePath, []byte(content), 0644)
}

func WriteOutputFile(sections []model.Section, filePath string) error {
	content := GenerateOutputMarkdown(sections)
	return os.WriteFile(filePath, []byte(content), 0644)
}

func GenerateInputMarkdown(sections []model.Section) string {
	var result strings.Builder

	for i, section := range sections {
		if i > 0 {
			result.WriteString("\n")
		}
		// Section header
		result.WriteString(fmt.Sprintf("## %s\n", section.Name))

		// Handle different section types
		switch section.Name {
		case model.SectionBacklog, model.SectionArchives:
			writeTasks(&result, section.Tasks, false)
		case model.SectionTodo, model.SectionDone:
			// These sections group tasks by date headers
			writeTasksWithDateHeaders(&result, section.Tasks)
		default:
			writeTasks(&result, section.Tasks, false)
		}
	}

	return result.String()
}

func GenerateOutputMarkdown(sections []model.Section) string {
	var result strings.Builder

	// Find Archives section
	var archiveTasks []model.Task
	for _, section := range sections {
		if section.Name == model.SectionArchives {
			archiveTasks = section.Tasks
			break
		}
	}

	for i, task := range archiveTasks {
		if i > 0 {

			result.WriteString("\n")
		}
		result.WriteString(taskToOutputMarkdown(task))
	}

	return result.String()
}

// writeTasks writes tasks in the output format
func taskToOutputMarkdown(task model.Task) string {
	var result strings.Builder

	// Merge project & title
	var title string

	if task.Project != "" {
		title += strings.ToUpper(task.Project) + " - "
	}

	title += task.Title

	// Title
	fmt.Fprintf(&result, "# %s\n", title)

	// Date range
	if task.StartDate != nil && task.EndDate != nil {
		if task.StartDate.Equal(*task.EndDate) {
			fmt.Fprintf(&result, "%s  \n", task.StartDate.Format("2006-01-02"))
		} else {
			fmt.Fprintf(&result, "%s - %s  \n",
				task.StartDate.Format("2006-01-02"),
				task.EndDate.Format("2006-01-02"))
		}
	}

	// Description
	if len(task.Description) > 0 {
		fmt.Fprintf(&result, "Desc:  \n")
		for _, desc := range task.Description {
			fmt.Fprintf(&result, "  %s  \n", desc)
		}
	}

	// Subtasks
	for _, subtask := range task.SubTasks {
		status := " "

		switch subtask.Status {
		case model.StatusDone:
			status = "x"
		case model.StatusInProgress:
			status = "-"
		}

		fmt.Fprintf(&result, "  - [%s] %s\n", status, subtask.Content)
	}

	return result.String()
}

// writeTasks writes tasks in the input format
func writeTasks(result *strings.Builder, tasks []model.Task, useHeaderDate bool) {
	for _, task := range tasks {
		writeTask(result, task, useHeaderDate)
	}
}

func writeTasksWithDateHeaders(result *strings.Builder, tasks []model.Task) {
	// Group tasks by date
	dateGroups := make(map[string][]model.Task)
	var dateOrder []string

	for _, task := range tasks {
		var dateKey string

		if task.StartDate != nil {
			dateKey = task.StartDate.Format("2006-01-02")
		} else {
			dateKey = "no-date"
		}

		if _, exists := dateGroups[dateKey]; !exists {
			dateOrder = append(dateOrder, dateKey)
		}
		dateGroups[dateKey] = append(dateGroups[dateKey], task)
	}

	for _, dateKey := range dateOrder {
		if dateKey != "no-date" {
			// Parse date back for formatting
			if date, err := time.Parse("2006-01-02", dateKey); err == nil {
				dayName := getDayName(date)
				fmt.Fprintf(result, "### %s - %s\n", dateKey, dayName)
			}
		}

		tasks := dateGroups[dateKey]
		for _, task := range tasks {
			writeTask(result, task, true)
		}

		result.WriteString("\n")
	}
}

func writeTask(result *strings.Builder, task model.Task, useHeaderDate bool) {
	// Task line with status and title
	status := " "
	switch task.Status {
	case model.StatusDone:
		status = "x"
	case model.StatusInProgress:
		status = "-"
	}

	// Build comment
	comment := buildTaskComment(task, useHeaderDate)

	if comment != "" {
		fmt.Fprintf(result, "- [%s] %s <!-- %s -->\n", status, task.Title, comment)
	} else {
		fmt.Fprintf(result, "- [%s] %s\n", status, task.Title)
	}

	// Write descriptions
	for _, desc := range task.Description {
		fmt.Fprintf(result, "  %s\n", desc)
	}

	// Write subtasks
	for _, subtask := range task.SubTasks {
		subtaskStatus := " "
		switch subtask.Status {
		case model.StatusDone:
			subtaskStatus = "x"
		case model.StatusInProgress:
			subtaskStatus = "-"
		}
		fmt.Fprintf(result, "  - [%s] %s\n", subtaskStatus, subtask.Content)
	}
}

func buildTaskComment(task model.Task, useHeaderDate bool) string {
	var parts []string

	// Add project
	if task.Project != "" {
		parts = append(parts, fmt.Sprintf("@%s", task.Project))
	}

	// Add ID
	if task.ID != "" {
		parts = append(parts, fmt.Sprintf("#%s", task.ID))
	}

	// Add dates only if we're not relying on the header date
	if !useHeaderDate {
		if dateStr := formatTaskDates(task.StartDate, task.EndDate); dateStr != "" {
			parts = append(parts, dateStr)
		}
	}

	return strings.Join(parts, "|")
}

// Helper functions
func formatTaskDates(startDate, endDate *time.Time) string {
	if startDate == nil && endDate == nil {
		return ""
	}

	if startDate != nil && endDate != nil {
		if startDate.Equal(*endDate) {
			return startDate.Format("2006-01-02")
		}
		return fmt.Sprintf("%s - %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	}

	if startDate != nil {
		return startDate.Format("2006-01-02")
	}

	return endDate.Format("2006-01-02")
}

func datesEqual(startDate, endDate, headerDate *time.Time) bool {
	if headerDate == nil || startDate == nil {
		return false
	}

	// Check if task dates match the header date
	return startDate.Equal(*headerDate) && (endDate == nil || endDate.Equal(*headerDate))
}

func getDayName(date time.Time) string {
	// Indonesian day names
	dayNames := map[time.Weekday]string{
		time.Sunday:    "Minggu",
		time.Monday:    "Senin",
		time.Tuesday:   "Selasa",
		time.Wednesday: "Rabu",
		time.Thursday:  "Kamis",
		time.Friday:    "Jum'at",
		time.Saturday:  "Sabtu",
	}
	return dayNames[date.Weekday()]
}
