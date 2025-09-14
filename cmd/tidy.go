package cmd

import (
	"fmt"
	"log"

	"github.com/ahmaruff/tada/internal/parser"
	"github.com/ahmaruff/tada/internal/processor"
	"github.com/ahmaruff/tada/internal/writer"
	"github.com/spf13/cobra"
)

var tidyCmd = &cobra.Command{
	Use:   "tidy [file]",
	Short: "Clean up and organize tasks",
	Long: `Clean up and organize tasks by consolidating data across sections.

This command:
1. Parse input file
2. Consolidate tasks (merge Backlog with Todo/Done data)
3. Optionally move completed Backlog tasks to Archives (--archive flag)
4. Update input file

Use --archive flag to move completed tasks from Backlog to Archives.`,
	Args: cobra.MaximumNArgs(1),
	Run:  runTidy,
}

var (
	tidyInputFile string
	tidyArchive   bool
	tidyDryRun    bool
	tidyVerbose   bool
)

func init() {
	tidyCmd.Flags().StringVarP(&tidyInputFile, "input", "i", "input.md", "Input markdown file")
	tidyCmd.Flags().BoolVarP(&tidyArchive, "archive", "a", false, "Move completed Backlog tasks to Archives")
	tidyCmd.Flags().BoolVar(&tidyDryRun, "dry-run", false, "Preview changes without applying them")
	tidyCmd.Flags().BoolVarP(&tidyVerbose, "verbose", "v", false, "Verbose output")
}

func runTidy(cmd *cobra.Command, args []string) {
	// Use positional argument if provided
	inputFile := tidyInputFile
	if len(args) > 0 {
		inputFile = args[0]
	}

	if tidyVerbose {
		fmt.Printf("Starting tada tidy with input: %s\n", inputFile)
		if tidyArchive {
			fmt.Println("Will move completed tasks to Archives")
		}
	}

	// 1. Parse input file
	if tidyVerbose {
		fmt.Println("1. Parsing input file...")
	}
	sections, err := parser.ParseFile(inputFile)
	if err != nil {
		log.Fatalf("Failed to parse input file: %v", err)
	}

	if tidyVerbose {
		fmt.Printf("   Parsed %d sections\n", len(sections))
		for _, section := range sections {
			fmt.Printf("   - %s: %d tasks\n", section.Name, len(section.Tasks))
		}
	}

	// Count tasks before consolidation
	var backlogBefore, completedBefore int
	for _, section := range sections {
		if section.Name == "Backlog" {
			backlogBefore = len(section.Tasks)
			for _, task := range section.Tasks {
				if task.Status == "StatusDone" {
					completedBefore++
				}
			}
			break
		}
	}

	// 2. Consolidate tasks
	if tidyVerbose {
		fmt.Println("\n2. Consolidating tasks...")
	}
	sections = processor.ConsolidateTasks(sections)

	// Count tasks after consolidation
	var backlogAfter, completedAfter int
	for _, section := range sections {
		if section.Name == "Backlog" {
			backlogAfter = len(section.Tasks)
			for _, task := range section.Tasks {
				if task.Status == "StatusDone" {
					completedAfter++
				}
			}
			break
		}
	}

	if tidyVerbose {
		fmt.Printf("   Backlog tasks: %d -> %d\n", backlogBefore, backlogAfter)
		fmt.Printf("   Completed tasks in Backlog: %d -> %d\n", completedBefore, completedAfter)
	}

	// 3. Optionally move completed tasks to Archives
	var movedCount int
	if tidyArchive {
		if tidyVerbose {
			fmt.Println("\n3. Moving completed tasks to Archives...")
		}
		sections = processor.MoveCompletedBacklogToArchives(sections)

		// Count moved tasks
		for _, section := range sections {
			if section.Name == "Backlog" {
				finalCompleted := 0
				for _, task := range section.Tasks {
					if task.Status == "StatusDone" {
						finalCompleted++
					}
				}
				movedCount = completedAfter - finalCompleted
				break
			}
		}

		if tidyVerbose {
			fmt.Printf("   Moved %d completed tasks to Archives\n", movedCount)
		}
	}

	if tidyDryRun {
		fmt.Printf("DRY RUN: Would consolidate %d tasks in Backlog\n", backlogAfter)
		if tidyArchive && movedCount > 0 {
			fmt.Printf("DRY RUN: Would move %d completed tasks to Archives\n", movedCount)
		}
		return
	}

	// 4. Update input file
	if tidyVerbose {
		fmt.Println("\n4. Updating input file...")
	}
	err = writer.WriteInputFile(sections, inputFile)
	if err != nil {
		log.Fatalf("Failed to write updated input file: %v", err)
	}

	// Summary output
	if tidyVerbose {
		fmt.Printf("   Updated input file: %s\n", inputFile)
		fmt.Println("\nTidy complete!")
	} else {
		fmt.Printf("Consolidated %d tasks", backlogAfter)
		if tidyArchive && movedCount > 0 {
			fmt.Printf(", moved %d to Archives", movedCount)
		}
		fmt.Printf(" in %s\n", inputFile)
	}
}
