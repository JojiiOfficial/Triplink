package main

import (
	"bufio"
	"fmt"
	"strings"
)

var yesInput = []string{"y", "yse", "yes", "ja", "si", "ofcourse", "ofc", "ys", "ye"}

func confirmInput(message string, reader *bufio.Reader) (bool, int) {
	i, txt := WaitForMessage(message, reader)
	return Contains(yesInput, strings.ToLower(txt)), i
}

//Contains returns true if a string slice has a given value
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

//WaitForMessage returns code (-1 = abort; 0 = empty; 1 = text) and the text
func WaitForMessage(question string, reader *bufio.Reader) (int, string) {
	fmt.Print(question)
	text, _ := reader.ReadString('\n')
	text = strings.ReplaceAll(text, "\n", "")
	if strings.ToLower(text) == "a" {
		return -1, ""
	}
	if len(text) > 0 {
		return 1, text
	}
	return 0, text
}
