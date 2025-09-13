package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ahmaruff/tada/internal/model"
)

type LineType int

const (
	LineUnknown       LineType = 0
	LineSectionHeader LineType = 1
	LineDateHeader    LineType = 2
	LineTask          LineType = 3
	LineDescription   LineType = 4
	LineSubtask       LineType = 5
)

var (
	sectionHeaderRegex = regexp.MustCompile(`^##\s(.+?)\s*$`)
	dateHeaderRegex    = regexp.MustCompile(`^###\s(\d{4}-\d{2}-\d{2})(?:\s.*)?$`)
	taskRegex          = regexp.MustCompile(`^-\s\[( |x|-)\]\s(.+?)(?:\s<!--(.+?)-->)?$`)
	subtaskRegex       = regexp.MustCompile(`^\s+-\s\[( |x|-)\]\s(.+)$`)
	descriptionRegex   = regexp.MustCompile(`^\s+.+$`)
)

func ParseFile(path string) ([]model.Section, error) {
	file, err := os.Open(path)
	if err != nil {
		return []model.Section{}, fmt.Errorf("failed to open file %s: %w", path, err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	return ParseContent(scanner)
}

func ParseContent(scanner *bufio.Scanner) ([]model.Section, error) {
	sections := []model.Section{}
	var currentSection *model.Section
	var currentTask *model.Task
	var currentDate *time.Time

	sectionNameList := map[string]model.SectionName{
		"Backlog":  model.SectionBacklog,
		"Archives": model.SectionArchives,
		"Todo":     model.SectionTodo,
		"Done":     model.SectionDone,
	}

	for scanner.Scan() {
		line := scanner.Text()
		lineType, extractedValue := checkLineType(line)

		switch lineType {
		case LineSectionHeader:
			// Save previous task before new section
			if currentSection != nil {
				if currentTask != nil {
					currentSection.Tasks = append(currentSection.Tasks, *currentTask)
				}
				sections = append(sections, *currentSection)
			}

			name, ok := sectionNameList[extractedValue]
			if !ok {
				name = model.SectionName(extractedValue)
			}

			currentSection = &model.Section{Name: name}
			currentTask = nil
		case LineDateHeader:
			// Save previous task before new date group
			if currentTask != nil && currentSection != nil {
				currentSection.Tasks = append(currentSection.Tasks, *currentTask)
				currentTask = nil
			}

			if date, err := time.Parse("2006-01-02", extractedValue); err == nil {
				currentDate = &date
			}
		case LineTask:
			// Save previous task
			if currentTask != nil && currentSection != nil {
				currentSection.Tasks = append(currentSection.Tasks, *currentTask)
			}

			// Parse new task, passing the current date
			task := parseTaskLine(line, currentDate)
			currentTask = &task

		case LineSubtask:
			if currentTask != nil {
				subtask := parseSubTaskLine(line)
				currentTask.SubTasks = append(currentTask.SubTasks, subtask)
			}
		case LineDescription:
			if currentTask != nil {
				currentTask.Description = append(currentTask.Description, extractedValue)
			}
		case LineUnknown:
			// ignore
		default:
			// ignore
		}

	}

	// The last task and section
	if currentTask != nil && currentSection != nil {
		currentSection.Tasks = append(currentSection.Tasks, *currentTask)
	}
	if currentSection != nil {
		sections = append(sections, *currentSection)
	}

	return sections, nil
}

func parseTaskLine(line string, date *time.Time) model.Task {
	// Single regex with groups: "- [status] title <!-- comment -->"
	matches := taskRegex.FindStringSubmatch(line)

	if len(matches) < 3 {
		return model.Task{} // Return empty if pattern doesn't match
	}

	// Parse status
	var status model.TaskStatus
	switch matches[1] {
	case " ":
		status = model.StatusTodo
	case "x":
		status = model.StatusDone
	case "-":
		status = model.StatusInProgress
	}

	// Extract title
	title := strings.TrimSpace(matches[2])

	// Parse comment if exists
	comment := ""
	if len(matches) > 3 {
		comment = strings.TrimSpace(matches[3])
	}

	// Parse comment for project, ID, and dates
	project, taskId, startDate, endDate := parseComment(comment)

	// Use fallback date if no dates found in comment
	if startDate == nil && date != nil {
		startDate = date
		endDate = date
	}

	return model.Task{
		ID:          taskId,
		Title:       title,
		Project:     project,
		Status:      status,
		StartDate:   startDate,
		EndDate:     endDate,
		Description: []string{},
		SubTasks:    []model.Subtask{},
	}
}

func parseSubTaskLine(line string) model.Subtask {
	// Single regex with groups: whitespace + "- [status] " + content
	matches := subtaskRegex.FindStringSubmatch(line)

	if len(matches) < 3 {
		return model.Subtask{} // Return empty if pattern doesn't match
	}

	var status model.TaskStatus
	switch matches[1] {
	case " ":
		status = model.StatusTodo
	case "x":
		status = model.StatusDone
	case "-":
		status = model.StatusInProgress
	}

	return model.Subtask{
		Status:  status,
		Content: strings.TrimSpace(matches[2]),
	}
}

func parseComment(comment string) (project string, taskId string, startDate *time.Time, endDate *time.Time) {
	if comment == "" {
		return
	}

	for part := range strings.SplitSeq(comment, "|") {
		part = strings.TrimSpace(part)

		if strings.HasPrefix(part, "@") {
			project = strings.TrimSpace(part[1:])
		} else if strings.HasPrefix(part, "#") {
			taskId = strings.TrimSpace(part[1:])
		} else if strings.Contains(part, " - ") {
			dates := strings.Split(part, " - ")
			if len(dates) == 2 {
				if start, err := time.Parse("2006-01-02", strings.TrimSpace(dates[0])); err == nil {
					startDate = &start
				}
				if end, err := time.Parse("2006-01-02", strings.TrimSpace(dates[1])); err == nil {
					endDate = &end
				}
			}
		} else if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, part); matched {
			// Single date: "2025-09-12"
			if date, err := time.Parse("2006-01-02", part); err == nil {
				startDate = &date
				endDate = &date
			}
		}

	}

	return
}

func checkLineType(line string) (LineType, string) {
	// Section header: ## Name
	if matches := sectionHeaderRegex.FindStringSubmatch(line); len(matches) > 1 {
		return LineSectionHeader, strings.TrimSpace(matches[1])
	}

	// Date header: ### YYYY-MM-DD (optional label)
	if matches := dateHeaderRegex.FindStringSubmatch(line); len(matches) > 1 {
		return LineDateHeader, matches[1]
	}

	// Task: - [ ] Something
	if taskRegex.MatchString(line) {
		return LineTask, ""
	}

	// Subtask:   - [ ] Something
	if subtaskRegex.MatchString(line) {
		return LineSubtask, ""
	}

	// Description: indented text
	if descriptionRegex.MatchString(line) {
		return LineDescription, strings.TrimSpace(line)
	}

	// Fallback
	return LineUnknown, ""
}
