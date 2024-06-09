package main

import (
	"net/http"
	"os"
	"strconv"
	"sync"
)

type Chunk struct {
	start int
	end int
}

type File struct {
	sync.Mutex
	ptr *os.File
}

func (f *File) WriteAt(b []byte, off int64) (n int, err error) {
	f.Lock()
	defer f.Unlock()
	return f.ptr.WriteAt(b, off)
}



func doPartialDownload(client *http.Client, file *File, url string, chunk Chunk) error {

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("Range", "bytes=" + strconv.Itoa(chunk.start) + "-" + strconv.Itoa(chunk.end))

	resp, err := client.Do(req)

	// TODO: status code check for 200 and 206

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	contents := make([]byte, chunk.end - chunk.start)

	

	etag := resp.Header.Get("ETag")
	println("ETag: ", etag)

	n, err := resp.Body.Read(contents)
	if err != nil {
		return err
	}
	println("Read: ", n)


	// if n != chunk.end - chunk.start {
	// 	return errors.New("read different bytes than expected")
	// }


	// println("Contents at ", chunk.start, " to ", chunk.end, " : ", string(contents))
	file.WriteAt(contents, int64(chunk.start))


	return nil
}


func downloadFile(filePath string, url string) error {
	// head request to get metadata
	resp, err := http.Head(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	headers := resp.Header

	file, err := os.Create(filePath)
	defer file.Close()

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
			MaxIdleConnsPerHost: 100,
		},
	}

	defer client.CloseIdleConnections()

	wg := sync.WaitGroup{}

	for i := 0; i < 5; i++ {
		chunk := Chunk {
			start: (i * length)/5,
			end: ((i+1) * length)/5,
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := doPartialDownload(client, &File{ptr: file}, url, chunk)
			if err != nil {
				panic(err)
			}
		}()
	}

	wg.Wait()


	return nil
	

}

func main() {
	x,_ := strconv.Atoi("-5")
	println(x)
	largeFileUrl := "https://raw.githubusercontent.com/json-iterator/test-data/master/large-file.json"
	fileName := "x.json"

	err := downloadFile(fileName, largeFileUrl)

	if err != nil {
		panic(err)
	}

	println("File downloaded")
}
