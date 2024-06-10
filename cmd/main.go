package main

import (
	"os"

	"github.com/sansmoraxz/godm"
	"github.com/spf13/cobra"
)


var rootCmd = &cobra.Command{
    Use:   "godm [URL] [output file]",
    Short: "Downloads a large file",
    Long:  `This application downloads a large file from a given URL.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var fileName = args[1]
		var largeFileUrl = args[0]
		if err := godm.DownloadFile(fileName, largeFileUrl, true); err != nil {
			println("Error downloading file:", err)
			os.Exit(1)
		}

	},
}

func main() {
	rootCmd.SetVersionTemplate("1.0.0\n")

    if err := rootCmd.Execute(); err != nil {
        println("Error executing command:", err)
        os.Exit(1)
    }
}