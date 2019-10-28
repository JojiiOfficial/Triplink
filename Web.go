package main

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
)

func request(url string, data []byte, ignoreCert bool) (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreCert},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return "", err
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(d), nil
}
