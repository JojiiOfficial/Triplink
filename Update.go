package main

import (
	"encoding/json"
	"fmt"
	"os"
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

		err := fetchIPs(config, configFile, argv.FetchAll)
		if err != nil {
			fmt.Println("Error fetching Update: " + err.Error())
		} else {
			fmt.Println("Update successfull")
		}

		return nil
	},
}

func fetchIPs(c *Config, configFile string, fetchAll bool) error {
	since := c.LastUpdate
	if fetchAll {
		since = 0
		//TODO delete all ipaddresses if full sync
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

	return nil
}

func blockIPs(ips []IPList) {
	fmt.Println(ips)
}
