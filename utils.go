package godm

import (
	"net/http"
	"strings"
)

func getHeaders(client *http.Client, url string, compress bool) (http.Header, error) {
	// head request to get metadata
	// Note: the uncompressed payload is considered for Content-Length

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}

	if compress {
		req.Header.Set("Accept-Encoding", strings.Join(supportedEncodings, ", "))
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp.Header, nil
}
