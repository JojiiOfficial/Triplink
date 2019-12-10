package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func request(url, file string, data []byte, ignoreCert bool) (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreCert},
	}
	client := &http.Client{Transport: tr}
	addFile := ""
	if strings.HasSuffix(url, "/") {
		addFile = url + file
	} else {
		addFile = url + "/" + file
	}
	resp, err := client.Post(addFile, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	response := strings.Trim(strings.ReplaceAll(string(d), "\n", ""), " ")

	isError, code, message := checkResponseErrors([]byte(response))
	if isError {
		if code == "error" {
			LogError(message)
		} else {
			fmt.Println("Got " + code + ": " + message)
		}
		return response, errors.New("Response error")
	}
	if !strings.HasSuffix(response, "}") && !strings.HasPrefix(response, "{") && !strings.HasSuffix(response, "]") && !strings.HasPrefix(response, "]") {
		fmt.Println(response)
		return response, errors.New("no json was returned")
	}
	return response, nil
}

func checkResponseErrors(response []byte) (isError bool, statuscode, errorMsg string) {
	var obj Status
	err := json.Unmarshal(response, &obj)
	if err != nil {
		return
	}
	if len(obj.StatusCode) > 0 && len(obj.StatusMessage) > 0 {
		isError = true
		statuscode = obj.StatusCode
		errorMsg = obj.StatusMessage
	}
	return
}
