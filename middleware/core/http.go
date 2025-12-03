package core

import (
	"bytes"
	"fmt"
	"net/http"
)

type Result struct {
	Data   []byte
	Status int
}

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
