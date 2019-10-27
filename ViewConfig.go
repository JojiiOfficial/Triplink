package main

import (
	"fmt"
	"os"

	"github.com/mkideal/cli"
)

type viewConfT struct {
	cli.Helper
}

var viewConfCMD = &cli.Command{
	Name:    "viewConfig",
	Aliases: []string{"vconf", "viewc", "showconf", "showconfig", "config", "conf", "confshow", "confview"},
	Desc:    "View configuration file",
	Argv:    func() interface{} { return new(viewConfT) },
	Fn: func(ctx *cli.Context) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Couldn't retrieve homeDir!")
			return nil
		}

		confFile := getConfFile(getConfPath(homeDir))
		_, err = os.Stat(confFile)
		if err != nil {
			fmt.Println("No config found. Nothing to do.")
			return nil
		}

		conf := readConfig(confFile)

		fmt.Println("-------- Configuration --------")
		fmt.Println("Host:\t\t", conf.Host)
		fmt.Println("LogFile:\t", conf.LogFile)
		fmt.Println("Token:\t\t", conf.Token)

		return nil
	},
}
