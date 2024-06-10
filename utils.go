package godm

import "net/http"

func getHeaders(client *http.Client, url string) (http.Header, error) {
	// head request to get metadata
	// Note: the uncompressed payload is considered for Content-Length
	// Use Accept-Encoding: gzip, deflate, br to get compressed payload size

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp.Header, nil
}
