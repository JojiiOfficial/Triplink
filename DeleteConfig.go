package main

import (
	"fmt"
	"os"

	"github.com/mkideal/cli"
)

type deleteConfT struct {
	cli.Helper
	ConfigName string `cli:"C,config" usage:"Secify the config to use" dft:"config.json"`
}

var deleteConfCMD = &cli.Command{
	Name:    "deleteConfig",
	Aliases: []string{"delconf", "deleteconfig", "dc", "dconf", "delc"},
	Desc:    "Delete configuration file",
	Argv:    func() interface{} { return new(deleteConfT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*deleteConfT)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Couldn't retrieve homeDir!")
			return nil
		}
		confPath := getConfPath(homeDir)
		confFile := getConfFile(confPath, argv.ConfigName)
		_, err = os.Stat(confFile)
		if err == nil {
			err := os.RemoveAll(confFile)
			if err != nil {
				fmt.Println("Couldn't delete configfile!")
				return nil
			}
			fmt.Println("Config deleted!")
		} else {
			fmt.Println("No config found. Nothing to do.")
		}
		return nil
	},
}
