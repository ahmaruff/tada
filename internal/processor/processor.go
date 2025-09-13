package processor

import "github.com/ahmaruff/tada/internal/model"

func ConsolidateTasks(sections []model.Section) []model.Section {
	// Build a map of task updates from Todo, Done, and Archives sections
	taskUpdates := make(map[string]*model.Task)

	for _, section := range sections {

		if section.Name == "Todo" || section.Name == "Done" || section.Name == "Archives" {
			for _, task := range section.Tasks {

				if task.ID != "" {
					// If we already have an update for this ID, merge the information
					if existingUpdate, exists := taskUpdates[task.ID]; exists {
						mergedTask := mergeTaskData(existingUpdate, &task)
						taskUpdates[task.ID] = mergedTask
					} else {
						// Make a copy of the task for the update map
						taskCopy := task
						taskUpdates[task.ID] = &taskCopy
					}
				}
			}
		}
	}

	// Create a new sections slice with updated data
	updatedSections := make([]model.Section, len(sections))

	for i, section := range sections {
		updatedSections[i] = model.Section{
			Name:  section.Name,
			Tasks: make([]model.Task, len(section.Tasks)),
		}

		for j, task := range section.Tasks {
			if section.Name == "Backlog" && task.ID != "" {
				// Update Backlog task if we have update data
				if updateData, exists := taskUpdates[task.ID]; exists {
					updatedSections[i].Tasks[j] = applyTaskUpdate(task, updateData)
				} else {
					// No update data, keep original
					updatedSections[i].Tasks[j] = task
				}
			} else {
				// Keep tasks from Todo, Done, Archives unchanged
				updatedSections[i].Tasks[j] = task
			}
		}
	}

	return updatedSections
}

// applyTaskUpdate applies update data to a Backlog task
func applyTaskUpdate(original model.Task, update *model.Task) model.Task {
	updated := original

	// Update status
	updated.Status = update.Status

	// Update dates
	if update.StartDate != nil {
		updated.StartDate = update.StartDate
	}
	if update.EndDate != nil {
		updated.EndDate = update.EndDate
	}

	// Merge descriptions
	updated.Description = mergeDescriptions(original.Description, update.Description)

	// Merge subtasks
	updated.SubTasks = mergeSubtasks(original.SubTasks, update.SubTasks)

	return updated
}

func mergeTaskData(existing *model.Task, new *model.Task) *model.Task {
	merged := *existing // Start with existing data

	// Update status - prioritize Done > InProgress > Todo
	if new.Status == model.StatusDone {
		merged.Status = model.StatusDone
	} else if new.Status == model.StatusInProgress && existing.Status != model.StatusDone {
		merged.Status = model.StatusInProgress
	}

	if new.StartDate != nil {
		if merged.StartDate == nil || new.StartDate.Before(*merged.StartDate) {
			merged.StartDate = new.StartDate
		}
	}

	if new.EndDate != nil {
		if merged.EndDate == nil || new.EndDate.After(*merged.EndDate) {
			merged.EndDate = new.EndDate
		}
	}

	merged.Description = mergeDescriptions(existing.Description, new.Description)

	merged.SubTasks = mergeSubtasks(existing.SubTasks, new.SubTasks)

	return &merged
}

func mergeDescriptions(existing []string, new []string) []string {
	if len(new) == 0 {
		return existing
	}

	if len(existing) == 0 {
		return new
	}

	descMap := make(map[string]bool)
	result := make([]string, 0, len(existing)+len(new))

	for _, desc := range existing {
		if desc != "" {
			descMap[desc] = true
			result = append(result, desc)
		}
	}

	for _, desc := range new {
		if desc != "" && !descMap[desc] {
			descMap[desc] = true
			result = append(result, desc)
		}
	}

	return result
}

// mergeSubtasks combines subtask slices, avoiding duplicates by content
func mergeSubtasks(existing []model.Subtask, new []model.Subtask) []model.Subtask {
	if len(new) == 0 {
		return existing
	}
	if len(existing) == 0 {
		return new
	}

	// Create a map to track subtasks by content
	subtaskMap := make(map[string]model.Subtask)

	// Add existing subtasks to map
	for _, subtask := range existing {
		if subtask.Content != "" {
			subtaskMap[subtask.Content] = subtask
		}
	}

	// Add or update subtasks from new data
	for _, newSubtask := range new {
		if newSubtask.Content != "" {
			if existingSubtask, exists := subtaskMap[newSubtask.Content]; exists {
				// Update existing subtask status if new one is "higher priority"
				if newSubtask.Status == model.StatusDone {
					existingSubtask.Status = model.StatusDone
				} else if newSubtask.Status == model.StatusInProgress && existingSubtask.Status != model.StatusDone {
					existingSubtask.Status = model.StatusInProgress
				}
				// Update the map with the modified subtask
				subtaskMap[newSubtask.Content] = existingSubtask
			} else {
				// Add new subtask
				subtaskMap[newSubtask.Content] = newSubtask
			}
		}
	}

	// Convert map back to slice
	result := make([]model.Subtask, 0, len(subtaskMap))
	for _, subtask := range subtaskMap {
		result = append(result, subtask)
	}

	return result
}
