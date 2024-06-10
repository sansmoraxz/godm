package main

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
)

type Chunk struct {
	start int
	end   int
	etag  string
	url   string
}


func(chunk *Chunk) doPartialDownload(client *http.Client, file *os.File) error {

	// TODO: handle with gzip compression

	req, _ := http.NewRequest("GET", chunk.url, nil)

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