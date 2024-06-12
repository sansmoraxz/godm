package main

import (
	"os"

	"github.com/sansmoraxz/godm"
	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	binPath, err := getCurrentBinaryName()
	if err != nil {
		binPath = "godm"
	}
	rootCmd := &cobra.Command{
		Use:   binPath,
		Short: "A download manager for large files",
	}

	rootCmd.SetVersionTemplate(version)
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(downloadCmd())
	rootCmd.AddCommand(logCmd())

	return rootCmd
}

func main() {
	if err := rootCmd().Execute(); err != nil {
		println("Error executing command. You may check the help command for more information.")
		println("Logs are stored in: ", godm.DefaultLogPath())
		os.Exit(1)
	}
}
