package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mkideal/cli"
)

type viewConfT struct {
	cli.Helper
	ConfigName string `cli:"C,config" usage:"Specify the config to use" dft:"config.json"`
	Verbose    int    `cli:"v,verbose" usage:"Specify how much logs should be displayed" dft:"0"`
}

var viewConfCMD = &cli.Command{
	Name:    "view",
	Aliases: []string{"v", "display"},
	Desc:    "View a configuration file",
	Argv:    func() interface{} { return new(viewConfT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*viewConfT)
		verboseLevel = argv.Verbose
		if len(strings.Trim(argv.ConfigName, " ")) == 0 {
			return errors.New("You need to specify a config")
		}
		confFile := getConfFile(getConfPath(getHome()), argv.ConfigName)
		if verboseLevel > 1 {
			fmt.Println("Config:", confFile)
		}
		_, err := os.Stat(confFile)
		if err != nil {
			return errors.New("Config \"" + argv.ConfigName + "\" not found")
		}

		conf := readConfig(confFile)

		fmt.Println("-------- Configuration --------")
		fmt.Println("File:\t\t", confFile)
		fmt.Println("Host:\t\t", conf.Host)
		fmt.Println("Token:\t\t", conf.Token)
		if os.Getuid() == 0 {
			fmt.Println("Auto rules:\t", conf.AutocreateIptables)
			fmt.Println("Last fetch:\t", parseTimeStamp(conf.Filter.Since))
		}

		if verboseLevel > 0 {
			var logadd string
			if len(conf.LogFile) > 0 {
				fmt.Println("LogFile:\t", conf.LogFile)
				logadd = "-f " + conf.LogFile + " "
			}
			if os.Getuid() == 0 {
				if conf.AutocreateIptables {
					logadd += "-R true"
				} else {
					logadd += "-R false"
				}
			}
			fmt.Println("\nRecreate this config:\ntriplink config create -t "+conf.Token, "-r", conf.Host, logadd)
		}

		return nil
	},
}
