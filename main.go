package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxP = 10

type Chunk struct {
	start int
	end   int
	etag  string
}

var logger *log.Logger

func doPartialDownload(client *http.Client, file *os.File, url string, chunk Chunk) error {

	// TODO: handle with gzip compression

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("Range", "bytes="+strconv.Itoa(chunk.start)+"-"+strconv.Itoa(chunk.end))

	resp, err := client.Do(req)

	// TODO: status code check esp. for dividing 200 and 206

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	etag := resp.Header.Get("ETag")
	logger.Println("P_ETag: ", etag)
	if etag != "" && etag != chunk.etag {
		return errors.New("ETag mismatch")
	}

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if len(contents)-1 != chunk.end-chunk.start {
		logger.Println("read different bytes than expected at " + strconv.Itoa(chunk.start) + "-" + strconv.Itoa(chunk.end) + " : " + strconv.Itoa(len(contents)))
	}

	n, err := file.Write(contents)
	if err != nil {
		return err
	}

	if n-1 != chunk.end-chunk.start {
		logger.Println("write different bytes than expected at " + strconv.Itoa(chunk.start) + "-" + strconv.Itoa(chunk.end) + " : " + strconv.Itoa(n))
	}

	return nil
}

func getHeaders(url string) (http.Header, error) {
	resp, err := http.Head(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp.Header, nil
}

func downloadFile(filePath string, url string) error {
	// head request to get metadata
	// Note: the uncompressed payload is considered for Content-Length
	// Use Accept-Encoding: gzip, deflate, br to get compressed payload size

	headers, err := getHeaders(url)
	if err != nil {
		return err
	}

	isAcceptRanges := headers.Get("Accept-Ranges") == "bytes"
	length, _ := strconv.Atoi(headers.Get("Content-Length")) // it will be 0 if not present
	etag := headers.Get("ETag")

	logger.Println("Content-Length: ", length)
	logger.Println("Accept-Ranges: ", isAcceptRanges)
	logger.Println("ETag: ", etag)

	// shared client
	client := &http.Client{
		// Timeout: time.Second * 10,
		Transport: &http.Transport{
			MaxIdleConns:        0,
			MaxIdleConnsPerHost: maxP * 2,
		},
	}

	defer client.CloseIdleConnections()

	chunkSize := 1024 * 1024 // 1MB

	toDownloadTracker := make(map[Chunk]bool)

	downBar := make([]bool, length/chunkSize+1)

	// download in parallel
	for i := 0; i < length/chunkSize+1; i++ {
		c := Chunk{
			start: i * chunkSize,
			end:   min((i+1)*chunkSize-1, length-1),
			etag:  etag,
		}
		toDownloadTracker[c] = false
	}
	wg := sync.WaitGroup{}

	// progress bar
	go func() {
		fmt.Println("Downloading...")
		for {
			fmt.Printf("\r")
			for _, v := range downBar {
				if v {
					fmt.Print("X")
				} else {
					fmt.Print("-")
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	sem := make(chan bool, maxP)

	for c := range toDownloadTracker {
		sem <- true
		wg.Add(1)
		go func(c Chunk) {
			defer func() {
				<-sem
				wg.Done()
			}()
			logger.Println("Downloading: ", c.start, c.end)
			file, err := os.Create(
				filePath + "." + strconv.Itoa(c.start) + "-" + strconv.Itoa(c.end) + ".part",
			)
			if err != nil {
				logger.Println("Error: ", err)
			}
			err = doPartialDownload(client, file, url, c)
			if err != nil {
				logger.Println("Error: ", err)
			}
			logger.Println("Downloaded: ", c.start, c.end)
			downBar[c.start/chunkSize] = true
		}(c)
	}

	wg.Wait()

	// reassemble the file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	for c := range toDownloadTracker {
		partFile, err := os.Open(
			filePath + "." + strconv.Itoa(c.start) + "-" + strconv.Itoa(c.end) + ".part",
		)
		
		if err != nil {
			return err
		}

		_, err = io.Copy(file, partFile)
		if err != nil {
			return err
		}

		partFile.Close()
		os.Remove(filePath + "." + strconv.Itoa(c.start) + "-" + strconv.Itoa(c.end) + ".part")

	}

	return nil
}

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
