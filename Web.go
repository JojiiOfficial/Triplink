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

func request(url, file string, data []byte, ignoreCert, showErrors bool) (string, bool, error) {
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
		return "", false, err
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", false, err
	}

	response := strings.Trim(strings.ReplaceAll(string(d), "\n", ""), " ")
	isError, isStatus, code, message := checkResponseErrors([]byte(response))
	if isError {
		if showErrors {
			if code == "error" {
				LogError(message)
			} else {
				fmt.Println("Got " + code + ": " + message)
			}
		}
		if isStatus {
			return response, isStatus, errors.New(message)
		} else {
			return response, isStatus, errors.New("Response error")
		}
	}
	if !strings.HasSuffix(response, "}") && !strings.HasPrefix(response, "{") && !strings.HasSuffix(response, "]") && !strings.HasPrefix(response, "]") {
		if showErrors {
			fmt.Println(response)
		}
		return response, false, errors.New("no json was returned")
	}
	return response, isStatus, nil
}

func checkResponseErrors(response []byte) (isError, isStatus bool, statuscode, errorMsg string) {
	status, err := responseToStatus(string(response))
	if err != nil {
		return
	}
	if len(status.StatusCode) > 0 && len(status.StatusMessage) > 0 {
		isStatus = true
		if status.StatusCode != "success" {
			isError = true
			statuscode = status.StatusCode
			errorMsg = status.StatusMessage
		}
	}
	return
}

func responseToStatus(resp string) (*Status, error) {
	var obj Status
	err := json.Unmarshal([]byte(resp), &obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}
