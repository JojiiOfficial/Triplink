package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
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

	scanner := bufio.NewScanner(logFile)
	scanner.Split(bufio.ScanLines)
	logs := ""
	for scanner.Scan() {
		line := scanner.Text()
		if len(strings.Trim(line, " ")) == 0 {
			continue
		}
		logEntry, err := parseLogEntry(line)
		if err != nil {
			fmt.Println("Couldn't parse le (" + err.Error() + "): " + line)
			continue
		}
		logs += line
		fmt.Println(*logEntry)
	}

	logFile.Close()

	if len(logs) > 0 {
		appendLogs(logFilePath+"_1", logs)
	}

	runCommand(nil, "echo -n > "+logFilePath)

}

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
	t, _ := time.Parse(time.Stamp, logItems[0]+" "+logItems[1]+" "+logItems[2])
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
