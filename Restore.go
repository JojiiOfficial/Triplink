package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/mkideal/cli"
)

type restoreT struct {
	cli.Helper
	RestoreIPtables bool   `cli:"t,iptables" usage:"Restore iptables" dft:"false"`
	RestoreIPset    bool   `cli:"s,ipset" usage:"Restore ipset" dft:"true"`
	ConfigName      string `cli:"C,config" usage:"Specify the config to use" dft:"config.json"`
	Verbose         int    `cli:"v,verbose" usage:"Specify how much logs should be displayed" dft:"0"`
}

var restoreCMD = &cli.Command{
	Name:    "restore",
	Aliases: []string{"res", "restore", "rest"},
	Desc:    "restore ipset and iptables",
	Argv:    func() interface{} { return new(restoreT) },
	Fn: func(ctx *cli.Context) error {
		if os.Getuid() != 0 {
			fmt.Println("You need to be root!")
			return nil
		}
		argv := ctx.Argv().(*restoreT)
		verboseLevel = argv.Verbose
		logStatus, configFile := createAndValidateConfigFile(argv.ConfigName)
		if logStatus != 1 {
			return errors.New("config not found")
		}
		restoreIPs(configFile, argv.RestoreIPset, argv.RestoreIPtables)
		return nil
	},
}

func restoreIPs(configFile string, restoreIPset, restoreIPtables bool) {
	configFolder, configfilename := path.Split(configFile)
	blocklistName := getBlocklistName(configfilename)
	iptablesFile := configFolder + "iptables_" + blocklistName + ".bak"
	ipsetFile := configFolder + "ipset_" + blocklistName + ".bak"

	if restoreIPset {
		if isIpsetInstalled(false) {
			stat, err := os.Stat(ipsetFile)
			if err != nil || stat.Size() == 0 {
				_, err = os.Create(ipsetFile)
				LogInfo("There is no ipset backup! Skipping")
			} else {
				runCommand(nil, "ipset flush "+blocklistName)
				_, err = runCommand(nil, "ipset restore < "+ipsetFile)
				if err != nil {
					LogError("Error restoring ipset: " + err.Error() + " -> \"" + "ipset restore < " + ipsetFile + "\"")
				} else {
					LogInfo("Successfully restored ipset")
				}
			}
		} else {
			LogInfo("IPset not installed, can't restore. Skipping")
		}
	}

	if restoreIPtables {
		stat, err := os.Stat(iptablesFile)
		if err != nil || stat.Size() == 0 {
			LogError("There is no iptables backup! Skipping")
		} else {
			_, err = runCommand(nil, "iptables-restore < "+iptablesFile)
			if err != nil {
				LogError("Error restoring iptables: " + err.Error() + "-> \"" + "iptables-restore < " + iptablesFile + "\"")
			} else {
				LogInfo("Successfully restored iptables")
			}
		}
	}
}
