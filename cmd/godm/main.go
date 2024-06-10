package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/sansmoraxz/godm"
	"github.com/spf13/cobra"
)

var download = &cobra.Command{
	Use:   "download [URL] [output file]",
	Short: "Downloads a large file",
	Long:  `This application downloads a large file from a given URL.`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var fileName = args[1]
		var largeFileUrl = args[0]
		var compress, err = cmd.Flags().GetBool("compress")
		if err != nil {
			println("Error getting flag value:", err)
			os.Exit(1)
		}

		if err := godm.DownloadFile(fileName, largeFileUrl, true, compress); err != nil {
			fmt.Printf("\n\nError downloading file: %v\n", err)
			os.Exit(1)
		}

	},
}

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Prints the log file path",
	Long:  `This command prints the path of the log file.`,
	Run: func(cmd *cobra.Command, args []string) {
		println("Log file path: ", godm.DefaultLogPath())
	},
}

func main() {
	rootCmd := &cobra.Command{
		Use:   getCurrentBinaryName(),
		Short: "A download manager for large files",
	}
	rootCmd.SetVersionTemplate("1.0.0\n")
	rootCmd.AddCommand(download)
	rootCmd.AddCommand(logCmd)

	download.Flags().BoolP("compress", "c", false, "Download with compression enabled. Default: false.")

	if err := rootCmd.Execute(); err != nil {
		println("Error executing command. You may check the help command for more information.")
		println("Logs are stored in: ", godm.DefaultLogPath())
		os.Exit(1)
	}
}

func getCurrentBinaryName() string {
	currentBinaryPath, err := os.Executable()
	if err != nil {
		println("Error getting current binary path:", err)
		os.Exit(1)
	}
	return strings.Split(currentBinaryPath, string(os.PathSeparator))[len(strings.Split(currentBinaryPath, string(os.PathSeparator)))-1]
}
