package main

import (
	"bufio"
	"os"

	"github.com/mkideal/cli"
)

type newConfT struct {
	cli.Helper
	LogFile    string `cli:"*f,file" usage:"Specify the file to read the logs from"`
	Host       string `cli:"*r,host" usage:"Specify the host to send the data to"`
	Token      string `cli:"*t,token" usage:"Specify the token required by uploading hosts"`
	Overwrite  bool   `cli:"o,overwrite" usage:"Overwrite current config" dft:"false"`
	Note       string `cli:"n,note" usage:"Sends a very short description"`
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
			LogInfo("Warning: Logfile doesn't exists!!")
		}
		logStatus, configFile := createAndValidateConfigFile(argv.ConfigName)
		if logStatus == 0 {
			createConf(configFile, argv, false)
		} else if logStatus == 1 {
			if argv.Overwrite {
				createConf(configFile, argv, true)
			} else {
				LogInfo("There is alread a config file! use -o to overwrite it!")
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
		Note:    argv.Note,
	}).save(configFile)
	if err == nil {
		if update {
			LogInfo("Config updated successfully!")
		} else {
			LogInfo("Config created successfully!")
		}
		LogInfo(configFile)
		if argv.SetFilter {
			createFilter(configFile)
		}
	} else {
		LogError("Error saving config File: " + err.Error())
	}
}
