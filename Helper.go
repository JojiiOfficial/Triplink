package main

import (
	"os"
	"os/exec"
)

func appendLogs(newf, logs string) {
	file, err := os.OpenFile(newf, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 755)
	if err != nil {
		panic(err)
	}
	_, err = file.WriteString(logs + "\n")
	if err != nil {
		panic(err)
	}
	file.Close()
}

func runCommand(errorHandler func(error, string), sCmd string) (outb string, err error) {
	out, err := exec.Command("su", "-c", sCmd).Output()
	output := string(out)
	if err != nil {
		if errorHandler != nil {
			errorHandler(err, sCmd)
		}
		return "", err
	}
	return output, nil
}
