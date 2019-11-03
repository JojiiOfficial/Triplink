package main

import (
	"fmt"
	"os"
	"path"

	"github.com/mkideal/cli"
)

type backupT struct {
	cli.Helper
	BackupIPtables bool `cli:"t,iptables" usage:"Update iptables" dft:"false"`
	BackupIPset    bool `cli:"s,ipset" usage:"Update ipset" dft:"true"`
}

var backupCMD = &cli.Command{
	Name:    "backup",
	Aliases: []string{"b", "bak", "backup"},
	Desc:    "backups ipset(-s) and (iptables with -t)",
	Argv:    func() interface{} { return new(backupT) },
	Fn: func(ctx *cli.Context) error {
		if os.Getuid() != 0 {
			fmt.Println("You need to be root!")
			return nil
		}
		argv := ctx.Argv().(*backupT)
		_, configFile := createAndValidateConfigFile("")
		backupIPs(configFile, argv.BackupIPset, argv.BackupIPtables)
		return nil
	},
}

func backupIPs(configFile string, updateIPset, updateIPtables bool) {
	configFolder, _ := path.Split(configFile)
	iptablesFile := configFolder + "iptables.bak"
	ipsetFile := configFolder + "ipset.bak"

	if updateIPtables {
		_, err := os.Stat(iptablesFile)
		if err != nil {
			_, err = os.Create(iptablesFile)
			if err != nil {
				LogError("Can't create backup file: " + iptablesFile)
			}
		}

		_, err = runCommand(nil, "iptables-save > "+iptablesFile)
		if err != nil {
			LogError("Couldn'd backup iptables: " + err.Error())
		} else {
			LogInfo("Iptables backup successfull")
		}
	}

	if updateIPset {
		_, err := os.Stat(ipsetFile)
		if err != nil {
			_, err = os.Create(ipsetFile)
			if err != nil {
				LogError("Can't create backup file: " + ipsetFile)
				return
			}
		}

		_, err = runCommand(nil, "ipset save blocklist > "+ipsetFile)
		if err != nil {
			LogError("Couldn'd backup ipset: " + err.Error())
		} else {
			LogInfo("Ipset backup successfull")
		}
	}
}
