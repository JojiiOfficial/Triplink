package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

func request(url string, data []byte) (string, error) {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return "", err
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(d), nil
}
