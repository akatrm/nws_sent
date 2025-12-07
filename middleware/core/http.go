// Package core provides a small HTTP helper used by Solr job workers
// to POST and GET JSON payloads to downstream services (for example
// the analytics engine). These functions are intentionally simple and
// synchronous; consider using a more robust HTTP client for production
// with retries and streaming support.
// TODO: add timeout handling, retries and context support
package core

import (
	"bytes"
	"fmt"
	"net/http"
)

// Result is a small wrapper for HTTP responses produced by
// PostRequest. It includes the returned status code and raw data.
type Result struct {
	Data   []byte
	Status int
}

// PostRequest performs a simple HTTP POST with a JSON content type.
// It returns a Result containing the response status and body or an
// error if the request or body read fails.
func PostRequest(url string, data []byte) (*Result, error) {
	res := &Result{}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return res, err
	}

	defer resp.Body.Close()

	bodyLength := resp.ContentLength

	var body []byte = make([]byte, bodyLength)
	total, err := resp.Body.Read(body)
	if err != nil {
		return res, err
	}

	res.Data = data
	res.Status = resp.StatusCode

	if total != int(bodyLength) {
		err := fmt.Errorf("incomplete read: expected %d bytes, got %d bytes", bodyLength, total)
		return res, err
	}

	return res, nil
}

// GetRequest performs a simple HTTP GET and returns the response body
// or an error. The helper currently reads the full body into memory.
func GetRequest(url string) ([]byte, error) {
	var respData []byte

	resp, err := http.Get(url)
	if err != nil {
		return respData, err
	}

	defer resp.Body.Close()

	bodyLength := resp.ContentLength
	var body []byte = make([]byte, bodyLength)
	total, err := resp.Body.Read(body)
	if err != nil {
		return respData, err
	}

	if total != int(bodyLength) {
		err := fmt.Errorf("incomplete read: expected %d bytes, got %d bytes", bodyLength, total)
		return respData, err
	}

	return body, nil
}
