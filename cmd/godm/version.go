package main

import "github.com/spf13/cobra"

func versionCmd() *cobra.Command {
	version := &cobra.Command{
		Use:   "version",
		Short: "Prints the version of the application",
		Long:  `This command prints the version of the application.`,
		Run: func(cmd *cobra.Command, args []string) {
			println("Version: ", version)
		},
	}
	return version
}
