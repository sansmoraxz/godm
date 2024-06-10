package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

const maxP = 10


var logger *log.Logger

var rootCmd = &cobra.Command{
    Use:   "godm [URL] [output file]",
    Short: "Downloads a large file",
    Long:  `This application downloads a large file from a given URL.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var fileName = args[1]
		var largeFileUrl = args[0]
		if err := downloadFile(fileName, largeFileUrl); err != nil {
			log.Println("Error downloading file:", err)
			os.Exit(1)
		}

	},
}

func main() {
	rootCmd.SetVersionTemplate("1.0.0\n")

	var logFile, err = os.Create(os.TempDir() + "/godm.log")
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	logger = log.New(logFile, "godm: ", log.LstdFlags)
	
    if err := rootCmd.Execute(); err != nil {
        log.Println("Error executing command:", err)
        os.Exit(1)
    }
}