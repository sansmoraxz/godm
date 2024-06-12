package main

import (
	"os"
	"runtime/debug"
	"strings"
)

var version = func() string {
	// only works with go version 1.18 and above
    if info, ok := debug.ReadBuildInfo(); ok {
        for _, setting := range info.Settings {
            if setting.Key == "vcs.revision" {
                return setting.Value
            }
        }
    }
    return ""
}()

func getCurrentBinaryName() string {
	currentBinaryPath, err := os.Executable()
	if err != nil {
		println("Error getting current binary path:", err)
		os.Exit(1)
	}
	return strings.Split(currentBinaryPath, string(os.PathSeparator))[len(strings.Split(currentBinaryPath, string(os.PathSeparator)))-1]
}

