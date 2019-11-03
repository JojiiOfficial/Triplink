package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/mkideal/cli"
)

type newConfT struct {
	cli.Helper
	LogFile    string `cli:"*f,file" usage:"Specify the file to read the logs from"`
	Host       string `cli:"*r,host" usage:"Specify the host to send the data to"`
	Token      string `cli:"*t,token" usage:"Specify the token required by uploading hosts"`
	Overwrite  bool   `cli:"o,overwrite" usage:"Overwrite current config" dft:"false"`
	SetFilter  bool   `cli:"F,filter" usage:"Specify to set the filter after creating the config" dft:"false"`
	ConfigName string `cli:"C,config" usage:"Specify the config to use" dft:"config.json"`
}

var createConfCMD = &cli.Command{
	Name:    "createConfig",
	Aliases: []string{"cc", "cconf", "createconf", "createconfig"},
	Desc:    "Create new configuration file",
	Argv:    func() interface{} { return new(newConfT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*newConfT)
		if os.Getuid() != 0 {
			reader := bufio.NewReader(os.Stdin)
			if y, _ := confirmInput("Warning! You are not root! Only root can report and fetch. Continue anyway? [y/n] ", reader); !y {
				return nil
			}
		}
		logFileExists := validateLogFile(argv.LogFile)
		if !logFileExists {
			fmt.Println("Warning: Logfile doesn't exists!!")
		}
		logStatus, configFile := createAndValidateConfigFile(argv.ConfigName)
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
	err := (&Config{
		Host:    argv.Host,
		LogFile: argv.LogFile,
		Token:   argv.Token,
	}).save(configFile)
	if err == nil {
		if update {
			fmt.Println("Config updated successfully!")
		} else {
			fmt.Println("Config created successfully!")
		}
		fmt.Println(configFile)
		if argv.SetFilter {
			createFilter(configFile)
		}
	} else {
		fmt.Println("Error saving config File: " + err.Error())
	}
}
