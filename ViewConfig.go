package main

import (
	"fmt"
	"os"

	"github.com/mkideal/cli"
)

type viewConfT struct {
	cli.Helper
	ConfigName string `cli:"C,config" usage:"Specify the config to use" dft:"config.json"`
}

var viewConfCMD = &cli.Command{
	Name:    "viewConfig",
	Aliases: []string{"vconf", "vc", "viewc", "showconf", "showconfig", "config", "conf", "confshow", "confview"},
	Desc:    "View a configuration file",
	Argv:    func() interface{} { return new(viewConfT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*viewConfT)

		confFile := getConfFile(getConfPath(getHome()), argv.ConfigName)
		fmt.Println(confFile)
		_, err := os.Stat(confFile)
		if err != nil {
			fmt.Println("No config found. Nothing to do.")
			return nil
		}

		conf := readConfig(confFile)

		fmt.Println("-------- Configuration --------")
		fmt.Println("Host:\t\t", conf.Host)
		fmt.Println("Token:\t\t", conf.Token)

		var logadd string
		if len(conf.LogFile) > 0 {
			fmt.Println("LogFile:\t", conf.LogFile)
			logadd = "-f " + conf.LogFile
		}
		fmt.Println("\nRecreate this config:\ntriplink cc -t "+conf.Token, "-r", conf.Host, logadd)

		return nil
	},
}
