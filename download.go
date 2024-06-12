package godm

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"sync"
	"time"
)

const (
	maxP      = 10
	chunkSize = 512 * 1024 // 1MB
)

type DownloadConfig struct {
	DisplayDownloadBar bool
	Compress           bool
}

type HeaderInfo struct {
	IsAcceptRanges bool
	Length         int
	ETag           string
	Encoding       string
}

func DownloadFile(filePath string, url string, config *DownloadConfig) error {

	// shared client
	client := &http.Client{
		// Timeout: time.Second * 10,
		Transport: &http.Transport{
			MaxIdleConns:        0,
			MaxIdleConnsPerHost: maxP * 2,
		},
	}
	defer client.CloseIdleConnections()

	headerInfo, err := getHeaders(client, url, config)
	if err != nil {
		return err
	}

	if headerInfo.Encoding == "" {
		config.Compress = false
	} else if slices.Contains(supportedEncodings, headerInfo.Encoding) {
		config.Compress = true
	} else {
		return errors.New("Unknown encoding: " + headerInfo.Encoding)
	}

	log.Info("Downloading: ", url)
	log.Info("Content-Length: ", headerInfo.Length)
	log.Info("Content-Encoding: ", headerInfo.Encoding)
	log.Info("Accept-Ranges: ", headerInfo.IsAcceptRanges)
	log.Info("ETag: ", headerInfo.ETag)

	toDownloadTracker := make(map[Chunk]bool)
	downBar := make([]bool, headerInfo.Length/chunkSize+1)

	if config.DisplayDownloadBar {
		go func() {
			for {
				// print download bar
				fmt.Printf("\r[")
				for i := 0; i < headerInfo.Length/chunkSize+1; i++ {
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
	for i := 0; i < headerInfo.Length/chunkSize+1; i++ {
		c := Chunk{
			start: i * chunkSize,
			end:   min((i+1)*chunkSize-1, headerInfo.Length-1),
			etag:  headerInfo.ETag,
			url:   url,
		}
		toDownloadTracker[c] = false
	}
	wg := sync.WaitGroup{}

	sem := make(chan bool, maxP)
	defer close(sem)
	wg.Add(headerInfo.Length/chunkSize + 1)

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
				err = c.doPartialDownload(client, file, config.Compress)
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
	
	if config.Compress {
		// intermediate archive file if compress is true
		archivePath := filePath + ".archive"
		if reassembledFile, err = os.Create(archivePath); err != nil {
			return err
		}
		defer func() {
			if err = reassembledFile.Close(); err != nil {
				log.Error("Error: ", err)
			}
			if err = os.Remove(archivePath); err != nil {
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

	if err = reassembleFile(reassembledFile, toDownloadTracker, partFileNameFn); err != nil {
		return err
	}

	log.Info("Reassembled file")

	if config.Compress {
		if err = decompressFile(reassembledFile, filePath, headerInfo.Encoding); err != nil {
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
