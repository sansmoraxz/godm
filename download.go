package godm

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
)

const maxP = 10

func DownloadFile(filePath string, url string) error {
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

	log.Info("Content-Length: ", length)
	log.Info("Accept-Ranges: ", isAcceptRanges)
	log.Info("ETag: ", etag)

	// shared client
	client := &http.Client{
		// Timeout: time.Second * 10,
		Transport: &http.Transport{
			MaxIdleConns:        0,
			MaxIdleConnsPerHost: maxP * 2,
		},
	}

	defer client.CloseIdleConnections()

	chunkSize := 512 * 1024 // 1MB

	toDownloadTracker := make(map[Chunk]bool)

	downBar := make([]bool, length/chunkSize+1)

	// download in parallel
	for i := 0; i < length/chunkSize+1; i++ {
		c := Chunk{
			start: i * chunkSize,
			end:   min((i+1)*chunkSize-1, length-1),
			etag:  etag,
			url:   url,
		}
		toDownloadTracker[c] = false
	}
	wg := sync.WaitGroup{}

	sem := make(chan bool, maxP)
	defer close(sem)
	wg.Add(length/chunkSize + 1)

	for c := range toDownloadTracker {
		sem <- true
		go func(c Chunk) {
			log.Trace("Acquired lock for: ", c.start, c.end)
			defer func() {
				<-sem
				wg.Done()
				log.Trace("Released lock for: ", c.start, c.end)
				log.Trace("Current sem: ", len(sem))
			}()
			file, err := os.Create(
				filePath + "." + strconv.Itoa(c.start) + "-" + strconv.Itoa(c.end) + ".part",
			)
			if err != nil {
				log.Error("Error: ", err)
			}
			defer file.Close()
			log.Info("Downloading: ", c.start, c.end)
			err = c.doPartialDownload(client, file)
			if err != nil {
				log.Error("Error: ", err)
			}
			log.Info("Downloaded: ", c.start, c.end)
			downBar[c.start/chunkSize] = true
		}(c)
	}

	wg.Wait()

	// reassemble the file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	log.Info("Reassembling file...")

	for c := range toDownloadTracker {
		partFile, err := os.Open(
			filePath + "." + strconv.Itoa(c.start) + "-" + strconv.Itoa(c.end) + ".part",
		)
		
		if err != nil {
			return err
		}

		// set file pointer to the start of the chunk
		_, err = file.Seek(int64(c.start), 0)
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
