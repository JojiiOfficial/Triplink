package main

import (
	"github.com/mkideal/cli"
)

type newConfT struct {
	cli.Helper
	Host       string `cli:"*r,host" usage:"Specify the host to send the data to"`
	Token      string `cli:"*t,token" usage:"Specify the token required by uploading hosts"`
	Overwrite  bool   `cli:"o,overwrite" usage:"Overwrite current config" dft:"false"`
	LogFile    string `cli:"f,file" usage:"Specify the file to read the logs from"`
	Note       string `cli:"n,note" usage:"Sends a very short description"`
	ConfigName string `cli:"C,config" usage:"Specify the config to use" dft:"config.json"`
}

var createConfCMD = &cli.Command{
	Name:    "createConfig",
	Aliases: []string{"cc", "cconf", "createconf", "createconfig"},
	Desc:    "Create a new configuration file",
	Argv:    func() interface{} { return new(newConfT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*newConfT)

		if len(argv.LogFile) > 0 {
			logFileExists := validateLogFile(argv.LogFile)
			if !logFileExists {
				LogInfo("Warning: Logfile doesn't exists!!")
			}
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
	}).save(configFile)
	if err == nil {
		if update {
			LogInfo("Config updated successfully!")
		} else {
			LogInfo("Config created successfully!")
		}
		LogInfo(configFile)
	} else {
		LogError("Error saving config File: " + err.Error())
	}
}
