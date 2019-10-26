package main

import (
	"fmt"

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
		logStatus, configFile := createAndValidateConfigFile(argv.LogFile)
		var config *Config
		if logStatus < 0 {
			return nil
		} else if logStatus == 0 {
			fmt.Println("Config empty. Using parameter as config. You can change them with <config>. Try 'twreporter help config' for more information.")
			if len(argv.Host) == 0 || len(argv.LogFile) == 0 || len(argv.Token) == 0 {
				fmt.Println("There is no config file! You have to set all arguments. Try 'twreporter help report'")
				return nil
			}
			config = &Config{
				Host:    argv.Host,
				LogFile: argv.LogFile,
				Token:   argv.Token,
			}
		} else {
			config = readConfig(configFile)
		}

		_ = config

		return nil
	},
}

//examle for reading the file line by line
// err := iptablesparser.ParseFileByLines("/var/log/Tripwire21", func(log *iptablesparser.LogEntry) {
// 	fmt.Println(*log)
// })
// if err != nil {
// 	panic(err)
// }
