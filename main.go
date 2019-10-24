package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	logFilePath := "/var/log/tst"

	if os.Getuid() != 0 {
		fmt.Println("You need to be root!")
		return
	}

	logFile, err := os.Open(logFilePath)
	if err != nil {
		fmt.Println("Error reading logfile!")
		return
	}
	defer logFile.Close()

	scanner := bufio.NewScanner(logFile)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		logEntry, err := parseLogEntry(line)
		if err != nil {
			fmt.Println("Couldn't parse le: " + line)
			continue
		}
		_ = logEntry
	}
}

//LogEntry a entry in log
type LogEntry struct {
}

func parseLogEntry(content string) (*LogEntry, error) {
	return nil, nil
}
