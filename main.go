package main

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

const maxP = 10

type Chunk struct {
	start int
	end int
}



func doPartialDownload(client *http.Client, file *os.File, url string, chunk Chunk) error {

	// TODO: handle with gzip compression

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("Range", "bytes=" + strconv.Itoa(chunk.start) + "-" + strconv.Itoa(chunk.end))

	resp, err := client.Do(req)

	// TODO: status code check esp. for dividing 200 and 206

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	etag := resp.Header.Get("ETag")
	println("P_ETag: ", etag)

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if len(contents)-1 != chunk.end - chunk.start {
		println("read different bytes than expected at " + strconv.Itoa(chunk.start) + "-" + strconv.Itoa(chunk.end) + " : " + strconv.Itoa(len(contents)))
	}

	n, err := file.Write(contents)
	if err != nil {
		return err
	}

	if n-1 != chunk.end - chunk.start {
		println("write different bytes than expected at " + strconv.Itoa(chunk.start) + "-" + strconv.Itoa(chunk.end) + " : " + strconv.Itoa(n))
	}


	return nil
}


func downloadFile(filePath string, url string) error {
	// head request to get metadata
	// Note: the uncompressed payload is considered for Content-Length
	// Use Accept-Encoding: gzip, deflate, br to get compressed payload size
	
	resp, err := http.Head(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	headers := resp.Header


	isAcceptRanges := headers.Get("Accept-Ranges") == "bytes"
	length, _ := strconv.Atoi(headers.Get("Content-Length")) // it will be 0 if not present
	etag := headers.Get("ETag")

	println("Content-Length: ", length)
	println("Accept-Ranges: ", isAcceptRanges)
	println("ETag: ", etag)


	// shared client
	client := &http.Client {
		// Timeout: time.Second * 10,
		Transport: &http.Transport {
			MaxIdleConns: 0,
			MaxIdleConnsPerHost: maxP*2,
		},
	}

	defer client.CloseIdleConnections()


	err = nil

	chunkSize := 1024 * 1024 // 1MB

	toDownloadTracker := make(map[Chunk]bool)

	// download in parallel
	for i := 0; i < length / chunkSize + 1; i++ {
		c := Chunk{
			start: i * chunkSize,
			end: min((i + 1) * chunkSize - 1, length - 1),
		}
		toDownloadTracker[c] = false
	}

	sem := make(chan bool, maxP)

	wg := sync.WaitGroup{}

	for c := range toDownloadTracker {
		sem <- true
		wg.Add(1)
		go func(c Chunk) {
			defer func() {
				<-sem
				wg.Done()
			}()
			println("Downloading: ", c.start, c.end)
			file, err := os.Create(
				filePath + "." + strconv.Itoa(c.start) + "-" + strconv.Itoa(c.end) + ".part",
			)
			if err != nil {
				println("Error: ", err)
			}
			err = doPartialDownload(client, file, url, c)
			if err != nil {
				println("Error: ", err)
			}
			println("Downloaded: ", c.start, c.end)
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
	}


	return nil
}

func main() {
	largeFileUrl := "https://raw.githubusercontent.com/json-iterator/test-data/master/large-file.json"
	fileName := largeFileUrl[strings.LastIndex(largeFileUrl, "/")+1:]

	err := downloadFile(fileName, largeFileUrl)

	if err != nil {
		panic(err)
	}

	println("File downloaded")
}
