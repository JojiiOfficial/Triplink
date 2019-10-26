package main

import (
	"fmt"

	iptablesparser "github.com/JojiiOfficial/Iptables-log-parser"
	"github.com/mkideal/cli"
)

type reportT struct {
	cli.Helper
	LogFile string `cli:"f,file" usage:"Specify the file to read the logs from"`
	Host    string `cli:"r,host" usage:"Specify the host to send the data to"`
	Token   string `cli:"t,token" usage:"Specify the token required by uploading hosts"`
}

var reportCMD = &cli.Command{
	Name:    "report",
	Aliases: []string{"r", "report", "repo"},
	Desc:    "Reports all changes",
	Argv:    func() interface{} { return new(reportT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*reportT)
		logStatus, configFile := createAndValidateConfigFile()
		var config *Config
		if logStatus < 0 {
			return nil
		} else if logStatus == 0 {
			fmt.Println("Config empty. Using parameter as config. You can change them with <config>. Try 'twreporter help config' for more information.")
			if len(argv.Host) == 0 || len(argv.LogFile) == 0 || len(argv.Token) == 0 {
				fmt.Println("There is no config file! You have to set all arguments. Try 'twreporter help report'")
				return nil
			}
			logFileExists := validateLogFile(argv.LogFile)
			if !logFileExists {
				fmt.Println("Logfile doesn't exists")
				return nil
			}
			config = &Config{
				Host:    argv.Host,
				LogFile: argv.LogFile,
				Token:   argv.Token,
			}
		} else {
			if len(argv.Host) != 0 && len(argv.LogFile) != 0 && len(argv.Token) != 0 {
				logFileExists := validateLogFile(argv.LogFile)
				if !logFileExists {
					fmt.Println("Logfile doesn't exists")
					return nil
				}
				fmt.Println("Using arguments instead of config!")
				config = &Config{
					Host:    argv.Host,
					LogFile: argv.LogFile,
					Token:   argv.Token,
				}
			} else if len(argv.Host) != 0 || len(argv.LogFile) != 0 || len(argv.Token) != 0 {
				fmt.Println("Arguments missing. Using config!")
				config = readConfig(configFile)
			} else {
				config = readConfig(configFile)
			}
		}

		logFileExists := validateLogFile(config.LogFile)
		if !logFileExists {
			fmt.Println("Logfile doesn't exists")
			return nil
		}

		fmt.Println("reading conf: ", config.LogFile)
		err := iptablesparser.ParseFileByLines(config.LogFile, func(log *iptablesparser.LogEntry) {
			fmt.Println(*log)
		})
		if err != nil {
			panic(err)
		}

		return nil
	},
}
