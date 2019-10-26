package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func request(url string, data []byte) (int, string) {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

	log.Println(result)
	return 0, ""
}
