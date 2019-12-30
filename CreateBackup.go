package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/mkideal/cli"
)

type backupT struct {
	cli.Helper
	BackupIPtables bool   `cli:"t,iptables" usage:"Update iptables" dft:"false"`
	BackupIPset    bool   `cli:"s,ipset" usage:"Update ipset" dft:"true"`
	ConfigName     string `cli:"C,config" usage:"Specify the config to use" dft:"config.json"`
	Verbose        int    `cli:"v,verbose" usage:"Specify how much logs should be displayed" dft:"0"`
}

var backupCMD = &cli.Command{
	Name:    "create",
	Aliases: []string{"b", "bak", "backup"},
	Desc:    "backups ipset(-s) and (iptables with -t)",
	Argv:    func() interface{} { return new(backupT) },
	Fn: func(ctx *cli.Context) error {
		if os.Getuid() != 0 {
			fmt.Println("You need to be root!")
			return nil
		}
		argv := ctx.Argv().(*backupT)
		verboseLevel = argv.Verbose
		if !argv.BackupIPtables && !argv.BackupIPset {
			LogInfo("nothing to do")
			return nil
		}
		logStatus, configFile := createAndValidateConfigFile(argv.ConfigName)
		if logStatus != 1 {
			return errors.New("config not found")
		}
		backupIPs(configFile, argv.BackupIPset, argv.BackupIPtables)
		return nil
	},
}

func backupIPs(configFile string, updateIPset, updateIPtables bool) {
	configFolder, configfilename := path.Split(configFile)
	blocklistName := getBlocklistName(configfilename)

	iptablesFile := configFolder + "iptables_" + blocklistName + ".bak"
	ipsetFile := configFolder + "ipset_" + blocklistName + ".bak"

	if updateIPtables {
		_, err := os.Stat(iptablesFile)
		if err != nil {
			_, err = os.Create(iptablesFile)
			if err != nil {
				LogError("Can't create backup file: " + iptablesFile)
			}
		} else {
			_, err := runCommand(nil, "iptables-save > "+iptablesFile)
			if err != nil {
				LogError("Couldn'd backup iptables: " + err.Error() + "-> \"" + "iptables-save > " + iptablesFile + "\"")
			} else {
				LogInfo("Iptables backup successfull")
			}
		}

	}

	if updateIPset {
		if isIpsetInstalled(false) {
			if !hasBlocklist(blocklistName) {
				LogInfo("Skipping ipset backup: no blocklist found")
				return
			}
			_, err := os.Stat(ipsetFile)
			if err != nil {
				_, err = os.Create(ipsetFile)
				if err != nil {
					LogError("Can't create backup file: " + ipsetFile)
					return
				}
			}
			_, err = runCommand(nil, "ipset save "+blocklistName+" > "+ipsetFile)
			if err != nil {
				LogError("Couldn'd backup ipset: " + err.Error() + "-> \"" + "ipset save " + blocklistName + " > " + ipsetFile + "\"")
			} else {
				LogInfo("Ipset backup successfull")
			}
		} else {
			LogInfo("You need to install ipset to backup ipset data. Skipping")
		}
	}
}
