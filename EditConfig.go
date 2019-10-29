package main

import (
	"fmt"
	"os"

	"github.com/mkideal/cli"
)

type editConfT struct {
	cli.Helper
	ConfigName string `cli:"C,config" usage:"Secify the config to use" dft:"config.json"`
	LogFile    string `cli:"f,file" usage:"Specify the file to read the logs from"`
	Host       string `cli:"r,host" usage:"Specify the host to send the data to"`
	Token      string `cli:"t,token" usage:"Specify the token required by uploading hosts"`
}

var editConfCMD = &cli.Command{
	Name:    "editConfig",
	Aliases: []string{"econf", "editconfig", "ec", "editc", "edconf"},
	Desc:    "Edit configuration file",
	Argv:    func() interface{} { return new(editConfT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*editConfT)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Couldn't retrieve homeDir!")
			return nil
		}
		confPath := getConfPath(homeDir)
		confFile := getConfFile(confPath, argv.ConfigName)
		_, err = os.Stat(confFile)
		if err == nil {
			realConf := readConfig(confFile)
			did := false
			if len(argv.LogFile) > 0 {
				logFileExists := validateLogFile(argv.LogFile)
				if !logFileExists {
					fmt.Println("Logfile doesn't exists")
					return nil
				}
				realConf.LogFile = argv.LogFile
				did = true
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
