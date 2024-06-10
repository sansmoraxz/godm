package godm

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	maxP      = 10
	chunkSize = 512 * 1024 // 1MB
)

func DownloadFile(filePath string, url string, displayDownloadBar bool, compress bool) error {
	// shared client
	client := &http.Client{
		// Timeout: time.Second * 10,
		Transport: &http.Transport{
			MaxIdleConns:        0,
			MaxIdleConnsPerHost: maxP * 2,
		},
	}
	defer client.CloseIdleConnections()

	headers, err := getHeaders(client, url, compress)
	if err != nil {
		return err
	}

	isAcceptRanges := headers.Get("Accept-Ranges") == "bytes"
	length, _ := strconv.Atoi(headers.Get("Content-Length")) // it will be 0 if not present
	etag := headers.Get("ETag")

	log.Info("Downloading: ", url)
	log.Info("Content-Length: ", length)
	log.Info("Accept-Ranges: ", isAcceptRanges)
	log.Info("ETag: ", etag)

	toDownloadTracker := make(map[Chunk]bool)
	downBar := make([]bool, length/chunkSize+1)

	if displayDownloadBar {
		go func() {
			for {
				// print download bar
				fmt.Printf("\r[")
				for i := 0; i < length/chunkSize+1; i++ {
					if downBar[i] {
						fmt.Printf("#")
					} else {
						fmt.Printf("-")
					}
				}
				fmt.Printf("]")
				time.Sleep(1 * time.Second)
			}
		}()
	}

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

	partFileNameFn := func(c Chunk) string {
		return filePath + "." + strconv.Itoa(c.start) + "-" + strconv.Itoa(c.end) + ".part"
	}

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
			file, err := os.Create(partFileNameFn(c))
			if err != nil {
				log.Error("Error: ", err)
			}
			defer file.Close()
			log.Info("Downloading: ", c.start, c.end)
			for !downBar[c.start/chunkSize] {
				err = c.doPartialDownload(client, file, compress)
				if err != nil {
					log.Error("Error for chunk: ", c.start, c.end, err)
				} else {
					log.Info("Downloaded: ", c.start, c.end)
					downBar[c.start/chunkSize] = true
				}
			}
		}(c)
	}

	wg.Wait()

	// reassemble the file

	var reassembledFile *os.File
	
	if compress {
		// intermediate gzip file if compress is true
		if reassembledFile, err = os.Create(filePath + ".gz"); err != nil {
			return err
		}
		defer func() {
			if err = reassembledFile.Close(); err != nil {
				log.Error("Error: ", err)
			}
			if err = os.Remove(filePath + ".gz"); err != nil {
				log.Error("Error: ", err)
			}
		}()
	} else {
		if reassembledFile, err = os.Create(filePath); err != nil {
			return err
		}
		defer reassembledFile.Close()
	}
	log.Info("Reassembling file...")

	reassembleFile(reassembledFile, toDownloadTracker, partFileNameFn)

	log.Info("Reassembled file")

	if compress {
		if err = decompressFile(reassembledFile, filePath); err != nil {
			return err
		}
	}

	return nil
}

func reassembleFile(mainFile *os.File, chunks map[Chunk]bool, partFileNameFn func(Chunk) string) error {
	for c := range chunks {
		err := func(c Chunk) error {
			partFile, err := os.Open(partFileNameFn(c))
			defer func() {
				if err = partFile.Close(); err != nil {
					log.Error("Error: ", err)
				}
				if err = os.Remove(partFileNameFn(c)); err != nil {
					log.Error("Error: ", err)
				}
			}()

			if err != nil {
				return err
			}

			if _, err = mainFile.Seek(int64(c.start), 0); err != nil {
				return err
			}

			if _, err = io.Copy(mainFile, partFile); err != nil {
				return err
			}
			log.Info("Reassembled: ", c.start, c.end)


			return nil
		}(c)
		if err != nil {
			return err
		}
	}
	return nil
}


func decompressFile(mainFile *os.File, filePath string) error {
	mainFile.Seek(0, 0)

	gr, err := gzip.NewReader(mainFile)
	if err != nil {
		return err
	}
	defer gr.Close()
	log.Info("Decompressing file...")
	fout, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fout.Close()
	if _, err = io.Copy(fout, gr); err != nil {
		return err
	}
	log.Info("Decompressed file")
	return nil
}
