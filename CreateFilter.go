package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

func createFilter(config string) {
	fmt.Println("Create a filter. Follow the steps. Keep empty to skip. Enter 'A' to exit.")
	reader := bufio.NewReader(os.Stdin)

	conf := readConfig(config)

	i, txt := WaitForMessage("\nWhich-min reason an IP must have to get blocked [1.0-3.0]: ", reader)
	if i == -1 {
		return
	}
	if i == 1 {
		minReason, err := strconv.ParseFloat(txt, 64)
		if minReason <= 0 || minReason > 3 {
			fmt.Println("Error! Choose a number between 1.0 and 3.0")
			return
		}
		if err == nil {
			conf.Filter.MinReason = minReason
		} else {
			fmt.Println("Not a float. Skipping.")
		}
	}

	i, txt = WaitForMessage("Which min amount of reports an IP must have to get blocked: ", reader)
	if i == -1 {
		return
	}
	if i == 1 {
		mineports, err := strconv.Atoi(txt)
		if err == nil {
			conf.Filter.MinReports = mineports
		} else {
			fmt.Println("Not an int. Skipping.")
		}
	}

	y, i := confirmInput("Allow proxies like tor-exit-nodes [y/n]: ", reader)
	if i == -1 {
		return
	}
	if i == 1 {
		if !y {
			conf.Filter.ProxyAllowed = -1
		} else {
			conf.Filter.ProxyAllowed = 0
		}
	}

	i, txt = WaitForMessage("Set a limit of max IPs you want to block [0=no limit]: ", reader)
	if i == -1 {
		return
	}
	if i == 1 {
		limit, err := strconv.Atoi(txt)
		if err == nil {
			if limit < 0 {
				fmt.Println("Can't be less than 0!")
				return
			}
			conf.Filter.MaxIPs = uint(limit)
		} else {
			fmt.Println("Not an int. Skipping.")
		}
	}

	y, i = confirmInput("Only validated IPs [y/n]: ", reader)
	if i == -1 {
		return
	}
	if i == 1 {
		if y {
			conf.Filter.OnlyValidatedIPs = -1
		} else {
			conf.Filter.OnlyValidatedIPs = 0
		}
	}

	err := conf.save(config)
	if err != nil {
		fmt.Println("Error saving config: " + err.Error())
	} else {
		fmt.Println("Config saved successfully")
	}

}
