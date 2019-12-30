package main

import (
	"fmt"
	"os"

	"github.com/mkideal/cli"
)

type editConfT struct {
	cli.Helper
	Host       string `cli:"r,host" usage:"Specify the host to send the data to"`
	Token      string `cli:"t,token" usage:"Specify the token required by uploading hosts"`
	ConfigName string `cli:"C,config" usage:"Specify the config to use" dft:"config.json"`
	LogFile    string `cli:"f,file" usage:"Specify the file to read the logs from. Use \"rem\" or \"remove\" to make it empty"`
	Ports      string `cli:"p,ports,port" usage:"Specify which ports will be blocked on IP-fetches" dft:"0-65535"`
	Verbose    int    `cli:"v,verbose" usage:"Specify how much logs should be displayed" dft:"0"`
}

var editConfCMD = &cli.Command{
	Name:    "editConfig",
	Aliases: []string{"econf", "editconfig", "ec", "editc", "edconf", "uc"},
	Desc:    "Edit a configuration file",
	Argv:    func() interface{} { return new(editConfT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*editConfT)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Couldn't retrieve homeDir!")
			return nil
		}
		if os.Getuid() != 0 && len(argv.Ports) > 0 && argv.Ports != "0-65535" {
			LogError("You can't specify ports. Only root is allowed to do that")
			return nil
		}
		confPath := getConfPath(homeDir)
		confFile := getConfFile(confPath, argv.ConfigName)
		_, err = os.Stat(confFile)
		if err == nil {
			realConf := readConfig(confFile)
			did := false
			if len(argv.LogFile) > 0 {
				if argv.LogFile == "del" || argv.LogFile == "rem" || argv.LogFile == "delete" || argv.LogFile == "remove" {
					did = true
					realConf.LogFile = ""
					fmt.Println("Removed logfile")
				} else {
					logFileExists := validateLogFile(argv.LogFile)
					if !logFileExists {
						fmt.Println("Warning!! Logfile doesn't exists!")
					}
					realConf.LogFile = argv.LogFile
					did = true
				}
			}
			if len(argv.Ports) > 0 && os.Getuid() == 0 {
				did = true
				ports, err := validatePortsParam(argv.Ports)
				if err != nil {
					return err
				}
				if deleteBlocklistIptableRules(getBlocklistName(argv.ConfigName)) {
					realConf.PortsToBlock = ports
					createIPtableRules(getBlocklistName(argv.ConfigName), realConf)
				} else {
					return nil
				}
			}
			if len(argv.Host) > 0 {
				realConf.Host = argv.Host
				did = true
			}
			if len(argv.Token) > 0 {
				if len(argv.Token) != 64 {
					fmt.Println("Your token is invalid!")
					return nil
				}
				realConf.Token = argv.Token
				realConf.Filter.Since = 0
				did = true
			}
			if !did {
				fmt.Println("Nothing to do!")
				return nil
			}
			err := realConf.save(confFile)
			if err == nil {
				fmt.Println("Config updated successfully!")
			} else {
				fmt.Println("Error saving config: " + err.Error())
			}
		} else {
			fmt.Println("Config not found. Nothing to do.")
		}
		return nil
	},
}
