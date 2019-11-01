package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func createFilter(config string) {
	fmt.Println("Create a filter. Follow the steps. Keep emty to skip. Enter 'A' to exit.")
	reader := bufio.NewReader(os.Stdin)

	conf := readConfig(config)

	fmt.Print("\nWhich-min reason an IP must have to get blocked [1.0-3.0]: ")
	text, _ := reader.ReadString('\n')
	text = strings.ReplaceAll(text, "\n", "")
	if len(text) > 0 {
		if text == "A" {
			return
		}
		minReason, err := strconv.ParseFloat(text, 64)
		if err == nil {
			conf.Filter.MinReason = minReason
		} else {
			fmt.Println("Not a float. Skipping.")
		}
	}

	fmt.Print("Which min amount of reports an IP must have to get blocked: ")
	text, _ = reader.ReadString('\n')
	text = strings.ReplaceAll(text, "\n", "")
	if len(text) > 0 {
		if text == "A" {
			return
		}
		mineports, err := strconv.Atoi(text)
		if err == nil {
			conf.Filter.MinReports = mineports
		} else {
			fmt.Println("Not an int. Skipping.")
		}
	}

	fmt.Print("Allow proxies like tor-exit-nodes [y/n]: ")
	text, _ = reader.ReadString('\n')
	text = strings.ReplaceAll(text, "\n", "")
	if len(text) > 0 {
		if text == "A" {
			return
		}
		conf.Filter.ProxyAllowed = 0
		if text == "no" || text == "n" || text == "0" || text == "false" {
			conf.Filter.ProxyAllowed = -1
		} else if text == "yes" || text == "y" || text == "1" || text == "true" {
			conf.Filter.ProxyAllowed = 0
		}
	}

	fmt.Print("Set a limit of max IPs you want to block [0=no limit]: ")
	text, _ = reader.ReadString('\n')
	text = strings.ReplaceAll(text, "\n", "")
	if len(text) > 0 {
		if text == "A" {
			return
		}
		limit, err := strconv.Atoi(text)
		if err == nil {
			conf.Filter.MaxIPs = uint(limit)
		} else {
			fmt.Println("Not an int. Skipping.")
		}
	}

	fmt.Print("Only validated IPs [y/n]: ")
	text, _ = reader.ReadString('\n')
	text = strings.ReplaceAll(text, "\n", "")
	if len(text) > 0 {
		if text == "A" {
			return
		}
		conf.Filter.OnlyValidatedIPs = 0
		if text == "no" || text == "n" || text == "0" || text == "false" {
			conf.Filter.OnlyValidatedIPs = 0
		} else if text == "yes" || text == "y" || text == "1" || text == "true" {
			conf.Filter.OnlyValidatedIPs = -1
		}
	}

	err := conf.save(config)
	if err != nil {
		fmt.Println("Error saving config: " + err.Error())
	} else {
		fmt.Println("Config saved successfully")
	}

}
