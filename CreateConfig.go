package main

import (
	"errors"
	"os"
	"strings"

	"github.com/mkideal/cli"
)

type newConfT struct {
	cli.Helper
	Host        string `cli:"r,host" usage:"Specify the host to send the data to"`
	Token       string `cli:"t,token" usage:"Specify the token required by uploading hosts"`
	Overwrite   bool   `cli:"o,overwrite" usage:"Overwrite current config" dft:"false"`
	LogFile     string `cli:"f,file" usage:"Specify the file to read the logs from"`
	ConfigName  string `cli:"C,config" usage:"Specify the config to use" dft:"config.json"`
	Ports       string `cli:"p,ports" usage:"Specify which ports will be blocked on IP-fetches" dft:"0-65535"`
	CreateRules bool   `cli:"R,create-rules" usage:"Auto create rules to block IPs" dft:"true"`
	Verbose     int    `cli:"v,verbose" usage:"Specify how much logs should be displayed" dft:"0"`
}

func (argv *newConfT) Validate(ctx *cli.Context) error {
	if len(strings.Trim(argv.Host, " ")) == 0 {
		return errors.New("Host missing. Use --host or -r to specify a host")
	}
	match, err := isURL(argv.Host)
	if err != nil{
		return err
	}
	if !match{
		return errors.New("Host must be an URL")
	}
	if len(strings.Trim(argv.Token, " ")) == 0 {
		return errors.New("Token missing. Use --token or -t to specify a token")
	}
	return nil
}

var createConfCMD = &cli.Command{
	Name:    "create",
	Aliases: []string{"c"},
	Desc:    "Create a new configuration file",
	Argv:    func() interface{} { return new(newConfT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*newConfT)
		verboseLevel = argv.Verbose
		if os.Getuid() != 0 {
			if len(argv.Ports) > 0 && argv.Ports != "0-65535" {
				LogError("You can't specify ports. Only root is allowed to do that")
				return nil
			}
			argv.CreateRules = false
		}

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
	ports, err := validatePortsParam(argv.Ports)
	if err != nil {
		LogError("Error parsing port param: " + err.Error())
		return
	}
	config := &Config{
		Host:               argv.Host,
		LogFile:            argv.LogFile,
		Token:              argv.Token,
		PortsToBlock:       ports,
		AutocreateIptables: argv.CreateRules,
	}

	err = config.save(configFile)
	if err == nil {
		pingSuccess := ping(config)
		if pingSuccess {
			if verboseLevel > 1 {
				LogInfo("Config successfully validated")
			}
			if update {
				LogInfo("Config updated successfully!")
			} else {
				LogInfo("Config created successfully!")
			}
			if verboseLevel > 0 {
				LogInfo(configFile)
			}
		} else {
			LogError("You can update your config using \"triplink config update -C " + argv.ConfigName + " -r <host> -t <token>\"")
		}
	} else {
		LogError("Error saving config File: " + err.Error())
	}
}
