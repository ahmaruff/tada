package cli

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "tada",
	Short: "A markdown-based task management tool",
	Long: `tada is a CLI tool for managing tasks in markdown format.
It consolidates task data across different sections and generates reports.

Complete documentation is available at https://github.com/ahmaruff/tada`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(genCmd)
	rootCmd.AddCommand(tidyCmd)
}
