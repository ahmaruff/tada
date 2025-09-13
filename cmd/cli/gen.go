package cli

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/ahmaruff/tada/internal/parser"
	"github.com/ahmaruff/tada/internal/processor"
	"github.com/ahmaruff/tada/internal/writer"
	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:   "gen [file]",
	Short: "Generate report from completed tasks",
	Long: `Generate a dated report from completed tasks.

This command runs the complete workflow:
1. Parse input file
2. Consolidate tasks (merge Backlog with Todo/Done data)
3. Move completed Backlog tasks to Archives
4. Generate report from Archives
5. Clear Archives
6. Update input file`,
	Args: cobra.MaximumNArgs(1),
	Run:  runGen,
}

var (
	genInputFile string
	genOutputDir string
	genDryRun    bool
	genVerbose   bool
)

func init() {
	genCmd.Flags().StringVarP(&genInputFile, "input", "i", "input.md", "Input markdown file")
	genCmd.Flags().StringVarP(&genOutputDir, "output", "o", ".", "Output directory for report")
	genCmd.Flags().BoolVar(&genDryRun, "dry-run", false, "Preview what would be processed without making changes")
	genCmd.Flags().BoolVarP(&genVerbose, "verbose", "v", false, "Verbose output")
}

func runGen(cmd *cobra.Command, args []string) {
	// Use positional argument if provided
	inputFile := genInputFile
	if len(args) > 0 {
		inputFile = args[0]
	}

	if genVerbose {
		fmt.Printf("Starting tada gen with input: %s\n", inputFile)
	}

	// 1. Parse input file
	if genVerbose {
		fmt.Println("1. Parsing input file...")
	}

	sections, err := parser.ParseFile(inputFile)
	if err != nil {
		log.Fatalf("Failed to parse input file: %v", err)
	}

	if genVerbose {
		fmt.Printf("   Parsed %d sections\n", len(sections))
		for _, section := range sections {
			fmt.Printf("   - %s: %d tasks\n", section.Name, len(section.Tasks))
		}
	}

	// 2. Consolidate tasks
	if genVerbose {
		fmt.Println("\n2. Consolidating tasks...")
	}

	sections = processor.ConsolidateTasks(sections)
	if genVerbose {
		fmt.Println("   Tasks consolidated")
	}

	// 3. Move completed Backlog tasks to Archives
	if genVerbose {
		fmt.Println("\n3. Moving completed tasks to Archives...")
	}

	sections = processor.MoveCompletedBacklogToArchives(sections)

	// Count archived tasks
	var archivedCount int
	for _, section := range sections {
		if section.Name == "Archives" {
			archivedCount = len(section.Tasks)
			break
		}
	}

	if genVerbose {
		fmt.Printf("   Moved to Archives: %d tasks\n", archivedCount)
	}

	if archivedCount == 0 {
		fmt.Println("No completed tasks found to archive. Report not generated.")
		return
	}

	if genDryRun {
		fmt.Printf("DRY RUN: Would generate report from %d archived tasks\n", archivedCount)
		if genVerbose {
			fmt.Println("Archived tasks:")
			for _, section := range sections {
				if section.Name == "Archives" {
					for i, task := range section.Tasks {
						dateStr := "no date"
						if task.StartDate != nil {
							dateStr = task.StartDate.Format("2006-01-02")
						}
						fmt.Printf("   %d. %s [%v] (%s)\n", i+1, task.Title, task.Status, dateStr)
					}
					break
				}
			}
		}
		return
	}

	// 4. Generate report with dynamic filename
	if genVerbose {
		fmt.Println("\n4. Generating report...")
	}

	// Find date range from Archives
	var earliestDate, latestDate *time.Time
	for _, section := range sections {
		if section.Name == "Archives" {
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
		filename := fmt.Sprintf("report_%s_%s.md",
			earliestDate.Format("2006-01-02"),
			latestDate.Format("2006-01-02"))
		outputFile = filepath.Join(genOutputDir, filename)
	} else {
		outputFile = filepath.Join(genOutputDir, "report.md")
	}

	err = writer.WriteOutputFile(sections, outputFile)
	if err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	fmt.Printf("Report generated: %s\n", outputFile)

	// 5. Clear Archives section
	if genVerbose {
		fmt.Println("\n5. Clearing Archives...")
	}
	sections = processor.ClearArchives(sections)

	// 6. Update input file
	if genVerbose {
		fmt.Println("\n6. Updating input file...")
	}
	err = writer.WriteInputFile(sections, inputFile)
	if err != nil {
		log.Fatalf("Failed to write updated input file: %v", err)
	}

	if genVerbose {
		fmt.Printf("   Updated input file: %s\n", inputFile)
		fmt.Println("\nProcessing complete!")
	} else {
		fmt.Printf("Input file updated: %s\n", inputFile)
	}
}
