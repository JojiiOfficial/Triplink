package main

import (
	"encoding/json"
	"fmt"

	"github.com/mkideal/cli"
)

type pingT struct {
	cli.Helper
	ConfigName string `cli:"C,config" usage:"Specify the config to use" dft:"config.json"`
	Host       string `cli:"r,host" usage:"Specify the host to send the data to"`
	Token      string `cli:"t,token" usage:"Specify the token required by uploading hosts"`
}

var pingCMD = &cli.Command{
	Name:    "ping",
	Aliases: []string{"p", "ping", "pin", "pi"},
	Desc:    "Checks connection to server is successful",
	Argv:    func() interface{} { return new(pingT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*pingT)

		logStatus, configFile := createAndValidateConfigFile(argv.ConfigName)
		var config *Config
		if logStatus < 0 {
			return nil
		} else if logStatus == 0 {
			fmt.Println(configEmptyError)
			if len(argv.Host) == 0 || len(argv.Token) == 0 {
				fmt.Println(noSuchConfigError)
				return nil
			}
			config = &Config{
				Host:  argv.Host,
				Token: argv.Token,
			}
		} else {
			fileConfig := readConfig(configFile)
			logFile := fileConfig.LogFile
			host := fileConfig.Host
			token := fileConfig.Token
			if len(argv.Host) > 0 {
				host = argv.Host
			}
			if len(argv.Token) > 0 {
				token = argv.Token
			}
			config = &Config{
				Host:    host,
				LogFile: logFile,
				Token:   token,
				Filter:  fileConfig.Filter,
			}
		}

		ping(config)
		return nil
	},
}

func ping(config *Config) bool {
	token := config.Token
	requestData := PingRequest{
		Token: token,
	}

	jsondata, err := json.Marshal(requestData)
	if err != nil {
		LogCritical("Error creating json:" + err.Error())
		return false
	}
	res, isStatus, err := request(config.Host, "ping", jsondata, true, false)
	if err != nil {
		LogError("Error validating config: " + err.Error())
		return false
	}
	if isStatus {
		status, _ := responseToStatus(res)
		if verboseLevel > 0 {
			LogInfo("Server response: " + status.StatusMessage)
		}
		return (status.StatusMessage == "success")
	}
	if len(res) > 30 {
		res = res[:30] + " [...]"
	}
	LogError("Got weird server response: " + res)
	return false
}
