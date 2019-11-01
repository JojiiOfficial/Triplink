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
	Desc:    "View configuration file",
	Argv:    func() interface{} { return new(viewConfT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*viewConfT)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Couldn't retrieve homeDir!")
			return nil
		}

		confFile := getConfFile(getConfPath(homeDir), argv.ConfigName)
		fmt.Println(confFile)
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
		fmt.Println("Filter: ")
		filter := conf.Filter
		fmt.Println("  min-Reason: \t", filter.MinReason, "1 = scanner, 2 spammer, 3 = bruteforcer")
		pa := "no"
		if filter.ProxyAllowed == 0 {
			pa = "yes"
		}
		fmt.Println("  Proxies allow:", pa)
		fmt.Println("  min-Reports: \t", filter.MinReports)
		fmt.Println("  maxIPs: \t", filter.MaxIPs)
		ov := "yes"
		if filter.OnlyValidatedIPs == 0 {
			ov = "no"
		}
		fmt.Println("  only valid:\t", ov)

		return nil
	},
}
