package main

import (
	"io"
	"net/http"
	"os"
)

func downloadFile(fileName string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	headers := resp.Header

	for k, v := range headers {
		println(k, v)
	}

	// Create the file
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
	

}

func main() {
	largeFileUrl := "https://raw.githubusercontent.com/json-iterator/test-data/master/large-file.json"
	fileName := "x.json"

	err := downloadFile(fileName, largeFileUrl)

	if err != nil {
		panic(err)
	}

	println("File downloaded")
}
