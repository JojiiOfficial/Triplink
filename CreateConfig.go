package main

import (
	"fmt"

	"github.com/mkideal/cli"
)

type newConfT struct {
	cli.Helper
	LogFile   string `cli:"*f,file" usage:"Specify the file to read the logs from"`
	Host      string `cli:"*r,host" usage:"Specify the host to send the data to"`
	Token     string `cli:"*t,token" usage:"Specify the token required by uploading hosts"`
	Overwrite bool   `cli:"o,overwrite" usage:"Overwrite current config" dft:"false"`
}

var createConfCMD = &cli.Command{
	Name:    "createConfig",
	Aliases: []string{"cc", "cconf", "createconf", "createconfig"},
	Desc:    "Create new configuration file",
	Argv:    func() interface{} { return new(newConfT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*newConfT)

		logStatus, configFile := createAndValidateConfigFile(argv.LogFile)
		if logStatus == 0 {
			createConf(configFile, argv, false)
		} else if logStatus == 1 {
			if argv.Overwrite {
				createConf(configFile, argv, true)
			} else {
				fmt.Println("There is alread a config file! use -o to overwrite it!")
			}
		}

		return nil
	},
}

func createConf(configFile string, argv *newConfT, update bool) {
	err := saveConfig(configFile, &Config{
		Host:    argv.Host,
		LogFile: argv.LogFile,
		Token:   argv.Token,
	})
	if err == nil {
		if update {
			fmt.Println("Config updated successfully!")
		} else {
			fmt.Println("Config created successfully!")
		}
	} else {
		fmt.Println("Error saving config File: " + err.Error())
	}
}
