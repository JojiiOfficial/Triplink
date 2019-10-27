package main

import (
	"fmt"
	"os"
	"path"

	"github.com/mkideal/cli"
)

type restoreT struct {
	cli.Helper
}

var restoreCMD = &cli.Command{
	Name:    "restore",
	Aliases: []string{"res", "restore"},
	Desc:    "restore ipset and iptables",
	Argv:    func() interface{} { return new(restoreT) },
	Fn: func(ctx *cli.Context) error {
		_, configFile := createAndValidateConfigFile()
		restoreIPs(configFile)
		return nil
	},
}

func restoreIPs(configFile string) {
	configFolder, _ := path.Split(configFile)
	iptablesFile := configFolder + "iptables.bak"
	ipsetFile := configFolder + "ipset.bak"

	_, err := os.Stat(ipsetFile)
	if err != nil {
		_, err = os.Create(ipsetFile)
		fmt.Println("Thereis no ipset backup!")
	} else {
		_, err = runCommand(nil, "ipset restore < "+ipsetFile)
		if err != nil {
			fmt.Println("Error restoring ipset:", err.Error())
		} else {
			fmt.Println("Successfully restored ipset")
		}
	}

	_, err = os.Stat(iptablesFile)
	if err != nil {
		fmt.Println("There is no iptables backup!")
	} else {
		_, err = runCommand(nil, "iptables-restore < "+iptablesFile)
		if err != nil {
			fmt.Println("Error restoring iptables:", err.Error())
		} else {
			fmt.Println("Successfully restored iptables")
		}
	}
}
