package main

import (
	"log"
	"os"
	"strings"
)

const maxP = 10


var logger *log.Logger


func main() {
	largeFileUrl := "https://raw.githubusercontent.com/json-iterator/test-data/master/large-file.json"
	fileName := largeFileUrl[strings.LastIndex(largeFileUrl, "/")+1:]

	var logFile, err = os.Create(os.TempDir() + "/godm.log")
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	logger = log.New(logFile, "godm: ", log.LstdFlags)

	// to disable logger
	// logger.SetOutput(io.Discard)

	err = downloadFile(fileName, largeFileUrl)

	if err != nil {
		panic(err)
	}

	logger.Println("File downloaded")
}
