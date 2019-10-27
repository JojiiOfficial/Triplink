package main

import (
	"fmt"
	"os"
	"path"

	"github.com/mkideal/cli"
)

type backupT struct {
	cli.Helper
}

var backupCMD = &cli.Command{
	Name:    "backup",
	Aliases: []string{"b", "bak", "backup"},
	Desc:    "backups ipset and iptables",
	Argv:    func() interface{} { return new(backupT) },
	Fn: func(ctx *cli.Context) error {
		_, configFile := createAndValidateConfigFile()
		backupIPs(configFile)
		return nil
	},
}

func backupIPs(configFile string) {
	configFolder, _ := path.Split(configFile)
	iptablesFile := configFolder + "iptables.bak"
	ipsetFile := configFolder + "ipset.bak"

	_, err := os.Stat(iptablesFile)
	if err != nil {
		_, err = os.Create(iptablesFile)
		if err != nil {
			fmt.Println("Can't create backup file: " + iptablesFile)
			return
		}
	}
	_, err = os.Stat(ipsetFile)
	if err != nil {
		_, err = os.Create(ipsetFile)
		if err != nil {
			fmt.Println("Can't create backup file: " + ipsetFile)
			return
		}
	}

	erro := false
	_, err = runCommand(nil, "iptables-save > "+iptablesFile)
	if err != nil {
		fmt.Println("Couldn'd backup iptables:", err.Error())
		erro = true
	}

	_, err = runCommand(nil, "ipset save blocklist > "+ipsetFile)
	if err != nil {
		fmt.Println("Couldn'd backup ipset:", err.Error())
		erro = true
	}

	if erro {
		fmt.Println("Error while backing up files")
	} else {
		fmt.Println("Successfully backed up iptables and ipset")
	}
}
