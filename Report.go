package main

import (
	"fmt"
	"os"

	"github.com/mkideal/cli"
)

type reportT struct {
	cli.Helper
	LogFile string `cli:"*f,file" usage:"Specify the file to read the logs from"`
	Host    string `cli:"t,host" usage:"Specify the host to send the data to"`
}

var reportCMD = &cli.Command{
	Name:    "report",
	Aliases: []string{"r", "report", "repo"},
	Desc:    "Reports all changes",
	Argv:    func() interface{} { return new(reportT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*reportT)
		logStatus := createFiles(argv.LogFile)
		if logStatus < 0 {
			return nil
		} else if logStatus == 0 {
			fmt.Println("Config empty. Using parameter as new config. You can change them with <config>. Try twreporter help config for more information.")
		}

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

func createFiles(logfile string) int {
	_, err := os.Stat(logfile)
	if err != nil {
		fmt.Println("Logfile doesn't exists")
		return 0
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Couldn't retrieve homeDir!")
		return -1
	}
	confPath := homeDir + "/" + ".tripwirereporter/"
	confFile := confPath + "conf.json"
	_, err = os.Stat(confPath)
	if err != nil {
		err = os.MkdirAll(confPath, os.ModePerm)
		if err != nil {
			fmt.Println("Couldn't create configpath")
			return -1
		}
		_, err = os.Create(confFile)
		if err != nil {
			fmt.Println("Couldn't create configfile")
			return -1
		}
	}
	_, err = os.Stat(confFile)
	if err != nil {
		_, err = os.Create(confFile)
		if err != nil {
			fmt.Println("Couldn't create configfile")
			return -1
		}
	}
	return 1
}
