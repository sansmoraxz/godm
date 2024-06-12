package main

import (
	"os"
	"path/filepath"
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

func getCurrentBinaryName() (string, error) {
	currentBinaryPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	currentBinaryPath, err = filepath.EvalSymlinks(currentBinaryPath)
	if ; err != nil {
		return "", err
	}
	currentBinaryPath = filepath.Base(currentBinaryPath)
	// whitespace is not allowed in the binary name
	if strings.Contains(currentBinaryPath, " ") {
		return "", os.ErrInvalid
	}
	return currentBinaryPath, nil
}
