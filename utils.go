package godm

import (
	"net/http"
	"strconv"
	"strings"
)

func getHeaders(client *http.Client, url string, config *DownloadConfig) (*HeaderInfo, error) {
	// head request to get metadata
	// Note: the uncompressed payload is considered for Content-Length

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}

	if config.Compress {
		req.Header.Set("Accept-Encoding", strings.Join(supportedEncodings, ", "))
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	l, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if  err != nil {
		return nil, err
	}

	return &HeaderInfo{
		IsAcceptRanges: resp.Header.Get("Accept-Ranges") == "bytes",
		Length:         l,
		ETag:           resp.Header.Get("ETag"),
		Encoding:       resp.Header.Get("Content-Encoding"),
	}, nil
}
