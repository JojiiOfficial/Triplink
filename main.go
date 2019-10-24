package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
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
			fmt.Println("Couldn't parse le (" + err.Error() + "): " + line)
			continue
		}
		fmt.Println(*logEntry)
	}
}

//LogEntry a entry in log
type LogEntry struct {
	Time                                           time.Time
	In, Out, Mac, Src, Dst, Len, TTL, ID, Protocol string
}

func parseLogEntry(content string) (*LogEntry, error) {
	if !strings.Contains(content, "Tripwire") {
		return nil, errors.New("not a tripwire log")
	}
	logItems := strings.Split(content, " ")
	entry := &LogEntry{}
	for _, val := range logItems {
		handleLogEntry(val, entry)
	}
	dateString := logItems[0] + " " + logItems[1] + " " + logItems[2]
	t, _ := time.Parse(time.Stamp, dateString)
	t = t.AddDate(time.Now().Year(), 0, 0)
	entry.Time = t
	return entry, nil
}

func handleLogEntry(data string, entry *LogEntry) {
	key, val, err := parseItem(data)
	if err != nil {
		return
	}
	switch key {
	case "IN":
		entry.In = val
	case "OUT":
		entry.Out = val
	case "MAC":
		entry.Mac = val
	case "SRC":
		entry.Src = val
	case "DST":
		entry.Dst = val
	case "LEN":
		entry.Len = val
	case "TTL":
		entry.TTL = val
	case "ID":
		entry.ID = val
	case "PROTO":
		entry.Protocol = val
	}
}

func parseItem(item string) (string, string, error) {
	if !strings.Contains(item, "=") {
		return "", "", errors.New("no valid item")
	}
	data := strings.Split(item, "=")
	if len(data) != 2 {
		return "", "", errors.New("no data given")
	}
	return data[0], data[1], nil
}
