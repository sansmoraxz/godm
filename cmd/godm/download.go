package main

import (
	"fmt"
	"os"

	"github.com/sansmoraxz/godm"
	"github.com/spf13/cobra"
)

func downloadCmd() *cobra.Command {
	download := &cobra.Command{
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

			if err := godm.DownloadFile(fileName, largeFileUrl, &godm.DownloadConfig{
				Compress: compress,
				DisplayDownloadBar: true,
			}); err != nil {
				fmt.Printf("\n\nError downloading file: %v\n", err)
				os.Exit(1)
			}

		},
	}
	download.Flags().BoolP("compress", "c", false, "Download with compression enabled. Default: false.")

	return download
}
