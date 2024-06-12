package main

import (
	"github.com/sansmoraxz/godm"
	"github.com/spf13/cobra"
)

func logCmd() *cobra.Command {
	log := &cobra.Command{
		Use:   "log",
		Short: "Prints the log file path",
		Long:  `This command prints the path of the log file.`,
		Run: func(cmd *cobra.Command, args []string) {
			println("Log file path: ", godm.DefaultLogPath())
		},
	}
	clearLog := &cobra.Command{
		Use:   "clear",
		Short: "Clears the log file",
		Long:  `This command clears the log file.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := godm.ClearLog(); err != nil {
				println("Error clearing log file:", err)
			}
		},
	}

	viewLog := &cobra.Command{
		Use:   "view",
		Short: "View the log file",
		Long:  `This command views the log file.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := godm.ViewLog(); err != nil {
				println("Error viewing log file:", err)
			}
		},
	}
	log.AddCommand(clearLog)
	log.AddCommand(viewLog)
	return log
}
