package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ahmaruff/tada/internal/model"
	"github.com/ahmaruff/tada/internal/parser"
	"github.com/ahmaruff/tada/internal/processor"
	"github.com/ahmaruff/tada/internal/writer"
)

func main() {
	fmt.Println("Starting tada processing...")

	// 1. Parse input file
	fmt.Println("1. Parsing input file...")
	sections, err := parser.ParseFile("example/input.md")
	if err != nil {
		log.Fatalf("Failed to parse input file: %v", err)
	}
	fmt.Printf("   Parsed %d sections\n", len(sections))

	// Print sections summary
	for _, section := range sections {
		fmt.Printf("   - %s: %d tasks\n", section.Name, len(section.Tasks))
	}

	// 2. Consolidate tasks (update Backlog with Todo/Done info)
	fmt.Println("\n2. Consolidating tasks...")
	sections = processor.ConsolidateTasks(sections)
	fmt.Println("   Tasks consolidated")

	// Print Backlog tasks status after consolidation
	for _, section := range sections {
		if section.Name == model.SectionBacklog {
			fmt.Printf("   Backlog after consolidation: %d tasks\n", len(section.Tasks))
			for i, task := range section.Tasks {
				if i < 3 { // Show first 3 tasks as sample
					fmt.Printf("     - %s [%v] (Project: %s, ID: %s)\n",
						task.Title, task.Status, task.Project, task.ID)
				}
			}
			if len(section.Tasks) > 3 {
				fmt.Printf("     ... and %d more\n", len(section.Tasks)-3)
			}
			break
		}
	}

	// 3. Move completed Backlog tasks to Archives
	fmt.Println("\n3. Moving completed tasks to Archives...")
	sections = processor.MoveCompletedBacklogToArchives(sections)

	// Show Archives after moving
	for _, section := range sections {
		if section.Name == model.SectionArchives {
			fmt.Printf("   Archives after moving: %d tasks\n", len(section.Tasks))
			for i, task := range section.Tasks {
				if i < 3 { // Show first 3 tasks as sample
					dateStr := "no date"
					if task.StartDate != nil {
						dateStr = task.StartDate.Format("2006-01-02")
					}
					fmt.Printf("     - %s [%v] (%s)\n", task.Title, task.Status, dateStr)
				}
			}
			if len(section.Tasks) > 3 {
				fmt.Printf("     ... and %d more\n", len(section.Tasks)-3)
			}
			break
		}
	}

	// 4. Generate report from Archives with dynamic filename
	fmt.Println("\n4. Generating report...")

	// Find date range from Archives
	var earliestDate, latestDate *time.Time
	for _, section := range sections {
		if section.Name == model.SectionArchives {
			for _, task := range section.Tasks {
				if task.StartDate != nil {
					if earliestDate == nil || task.StartDate.Before(*earliestDate) {
						earliestDate = task.StartDate
					}
					if task.EndDate != nil {
						if latestDate == nil || task.EndDate.After(*latestDate) {
							latestDate = task.EndDate
						}
					} else {
						if latestDate == nil || task.StartDate.After(*latestDate) {
							latestDate = task.StartDate
						}
					}
				}
			}
		}
	}

	// Generate filename
	var outputFile string
	if earliestDate != nil && latestDate != nil {
		outputFile = fmt.Sprintf("example/report_%s_%s.md",
			earliestDate.Format("2006-01-02"),
			latestDate.Format("2006-01-02"))
	} else {
		outputFile = "example/report.md"
	}

	err = writer.WriteOutputFile(sections, outputFile)
	if err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}
	fmt.Printf("   Report generated: %s\n", outputFile)

	// // 5. Clear Archives section
	// fmt.Println("\n5. Clearing Archives...")
	// sections = processor.ClearArchives(sections)
	//
	// // Verify Archives is empty
	// for _, section := range sections {
	// 	if section.Name == model.SectionArchives {
	// 		fmt.Printf("   Archives after clearing: %d tasks\n", len(section.Tasks))
	// 		break
	// 	}
	// }

	// 6. Update input file (now without completed tasks in Backlog)
	fmt.Println("\n6. Updating input file...")
	err = writer.WriteInputFile(sections, "example/input_updated.md")
	if err != nil {
		log.Fatalf("Failed to write updated input file: %v", err)
	}
	fmt.Println("   Updated input saved as: example/input_updated.md")

	// Final summary
	fmt.Println("\nProcessing complete!")
	fmt.Println("Files generated:")
	fmt.Printf("   - %s (report)\n", outputFile)
	fmt.Println("   - example/input_updated.md (updated input)")

	// Show final Backlog status
	for _, section := range sections {
		if section.Name == model.SectionBacklog {
			completedCount := 0
			todoCount := 0
			inProgressCount := 0

			for _, task := range section.Tasks {
				switch task.Status {
				case model.StatusDone:
					completedCount++
				case model.StatusInProgress:
					inProgressCount++
				default:
					todoCount++
				}
			}

			fmt.Printf("\nFinal Backlog status:\n")
			fmt.Printf("   - Todo: %d tasks\n", todoCount)
			fmt.Printf("   - In Progress: %d tasks\n", inProgressCount)
			fmt.Printf("   - Completed: %d tasks (should be 0 after moving to archives)\n", completedCount)
			break
		}
	}

}
