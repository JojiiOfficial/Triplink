package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mkideal/cli"
)

type updateConfT struct {
	cli.Helper
	Host     string `cli:"r,host" usage:"Specify the host to send the data to"`
	Token    string `cli:"t,token" usage:"Specify the token required by uploading hosts"`
	FetchAll bool   `cli:"a,all" usage:"Fetches everything"`
}

var updateCMD = &cli.Command{
	Name:    "update",
	Aliases: []string{"u", "upd", "update"},
	Desc:    "updates the ipset",
	Argv:    func() interface{} { return new(updateConfT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*updateConfT)
		if os.Getuid() != 0 {
			fmt.Println("You need to be root!")
			return nil
		}
		if !checkCommands() {
			return nil
		}

		setupIPset()

		logStatus, configFile := createAndValidateConfigFile()
		var config *Config
		if logStatus < 0 {
			return nil
		} else if logStatus == 0 {
			fmt.Println("Config empty. Using parameter as config. You can change them with <config>. Try 'twreporter help config' for more information.")
			if len(argv.Host) == 0 || len(argv.Token) == 0 {
				fmt.Println("There is no config file! You have to set all arguments. Try 'twreporter help report'")
				return nil
			}
			config = &Config{
				Host:  argv.Host,
				Token: argv.Token,
			}
		} else {
			if len(argv.Host) != 0 && len(argv.Token) != 0 {
				fmt.Println("Using arguments instead of config!")
				config = &Config{
					Host:  argv.Host,
					Token: argv.Token,
				}
			} else if len(argv.Host) != 0 || len(argv.Token) != 0 {
				fmt.Println("Arguments missing. Using config!")
				config = readConfig(configFile)
			} else {
				config = readConfig(configFile)
			}
		}

		err := FetchIPs(config, configFile, argv.FetchAll)
		if err != nil {
			fmt.Println("Error fetching Update: " + err.Error())
		}

		return nil
	},
}

//FetchIPs fetches IPs and puts them into a blocklist
func FetchIPs(c *Config, configFile string, fetchAll bool) error {
	since := c.LastUpdate
	if fetchAll {
		since = 0
		flusIPset()
	}
	requestData := FetchRequest{
		Token: c.Token,
		Filter: FetchFilter{
			Since: since,
		},
	}
	js, err := json.Marshal(requestData)
	if err != nil {
		return err
	}

	data, err := request(c.Host+"/fetch", js)
	data = strings.ReplaceAll(data, "\n", "")
	if err != nil || data == "\"[]\"" {
		if data == "\"[]\"" {
			fmt.Println("Nothing to do (updating)")
		}
		return err
	}

	var fetchresponse FetchResponse
	err = json.Unmarshal([]byte(data), &fetchresponse)
	if err != nil {
		return err
	}

	c.LastUpdate = fetchresponse.CurrentTimestamp
	c.save(configFile)

	blockIPs(fetchresponse.IPs)
	backupIPs(configFile, true, false)
	return nil
}

func blockIPs(ips []IPList) {
	addCount := 0
	remCount := 0
	for _, ip := range ips {
		if ip.Deleted == 1 {
			if ipsetRemoveIP(ip.IP) {
				remCount++
			}
		} else {
			if ipsetAddIP(ip.IP) {
				addCount++
			}
		}
	}
	if activateIPset() {
		fmt.Println("Successfully added "+strconv.Itoa(addCount), "and removed "+strconv.Itoa(remCount), "IPs")
	}
}

func activateIPset() bool {
	if iptableHasRule() {
		return true
	}
	_, err := runCommand(nil, "iptables -A INPUT -m set --match-set blocklist src -j DROP")
	if err != nil {
		fmt.Println("Couldn't activate iptable set. Blocking might be unavailable")
		return false
	}
	return true
}

func flusIPset() {
	runCommand(nil, "ipset flush blocklist")
}

func iptableHasRule() bool {
	_, err := runCommand(nil, "iptables -C INPUT -m set --match-set blocklist src -j DROP")
	return err == nil
}

func ipsetAddIP(ip string) bool {
	valid, _ := isIPValid(ip)
	if valid {
		_, err := runCommand(nil, "ipset add blocklist "+ip)
		return err == nil
	}
	return false
}

func ipsetRemoveIP(ip string) bool {
	valid, _ := isIPValid(ip)
	if valid {
		_, err := runCommand(nil, "ipset del blocklist "+ip)
		return err == nil
	}
	return false
}

func checkCommands() bool {
	_, err := runCommand(nil, "ipset help")
	if err != nil {
		fmt.Println("You need to install 'ipset' to run this command!")
		return false
	}
	return true
}

func hasBlocklist() bool {
	_, err := runCommand(nil, "ipset list blocklist")
	return err == nil
}

func createBlocklist() bool {
	_, err := runCommand(nil, "ipset create blocklist nethash")
	return err == nil
}

func setupIPset() {
	if !hasBlocklist() {
		if !createBlocklist() {
			fmt.Println("Couldn't create blocklist! Exiting")
			os.Exit(1)
		}
	}
}
