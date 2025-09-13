package processor

import (
	"sort"

	"github.com/ahmaruff/tada/internal/model"
)

func MoveCompletedBacklogToArchives(sections []model.Section) []model.Section {
	result := make([]model.Section, len(sections))

	var completedTasks []model.Task

	for i, section := range sections {
		result[i] = model.Section{
			Name:  section.Name,
			Tasks: make([]model.Task, 0),
		}

		switch section.Name {
		case model.SectionBacklog:
			// Split tasks: completed go to archives, incomplete stay in backlog
			for _, task := range section.Tasks {
				if task.Status == model.StatusDone {
					completedTasks = append(completedTasks, task)
				} else {
					result[i].Tasks = append(result[i].Tasks, task)
				}
			}
		case model.SectionArchives:
			// Add existing archive tasks plus new completed tasks
			result[i].Tasks = append(result[i].Tasks, section.Tasks...)
			result[i].Tasks = append(result[i].Tasks, completedTasks...)

			sortTasksByDate(result[i].Tasks)
		default:
			// Keep other sections unchanged
			result[i].Tasks = section.Tasks
		}
	}

	return result
}

func ClearArchives(sections []model.Section) []model.Section {
	result := make([]model.Section, len(sections))

	for i, section := range sections {
		result[i] = model.Section{
			Name: section.Name,
		}

		if section.Name == model.SectionArchives {
			// Clear all tasks from Archives
			result[i].Tasks = []model.Task{}
		} else {
			// Keep other sections unchanged
			result[i].Tasks = section.Tasks
		}
	}

	return result
}

func sortTasksByDate(tasks []model.Task) {
	sort.Slice(tasks, func(i, j int) bool {
		// Handle nil dates - put them at the end
		if tasks[i].StartDate == nil && tasks[j].StartDate == nil {
			return false // maintain original order for tasks without dates
		}

		if tasks[i].StartDate == nil {
			return false
		}

		if tasks[j].StartDate == nil {
			return true
		}

		// Sort by start date descending (newer first)
		return tasks[i].StartDate.After(*tasks[j].StartDate)
	})
}
