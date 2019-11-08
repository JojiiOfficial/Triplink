package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mkideal/cli"
)

type reportT struct {
	cli.Helper
	LogFile          string `cli:"f,file" usage:"Specify the file to read the logs from"`
	Host             string `cli:"r,host" usage:"Specify the host to send the data to"`
	Token            string `cli:"t,token" usage:"Specify the token required by uploading hosts"`
	Note             string `cli:"n,note" usage:"Sends a very short description"`
	DoUpdate         bool   `cli:"u,update" usage:"Specify if the client should update after the report" dft:"false"`
	UpdateEverything bool   `cli:"a,all" usage:"Specify if the client should update everything if update is set" dft:"false"`
	CustomIPs        string `cli:"c,custom" usage:"Report a custom IPset separated by semicolon and comma (eg: \"ip,port,count;ip2,port,count\")"`
	IgnoreCert       bool   `cli:"i,ignorecert" usage:"Ignore invalid certs" dft:"false"`
	ConfigName       string `cli:"C,config" usage:"Specify the config to use" dft:"config.json"`
}

var reportCMD = &cli.Command{
	Name:    "report",
	Aliases: []string{"r", "report", "repo"},
	Desc:    "Reports all changes",
	Argv:    func() interface{} { return new(reportT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*reportT)
		if os.Getuid() != 0 {
			fmt.Println("You need to be root!")
			return nil
		}
		logStatus, configFile := createAndValidateConfigFile(argv.ConfigName)
		var config *Config
		if logStatus < 0 {
			return nil
		} else if logStatus == 0 {
			fmt.Println("Config empty. Using parameter as config. You can change them with <config>. Try 'twreporter help config' for more information.")
			if len(argv.Host) == 0 || len(argv.LogFile) == 0 || len(argv.Token) == 0 {
				fmt.Println("There is no such config file! You have to set all arguments. Try 'twreporter help report'")
				return nil
			}
			logFileExists := validateLogFile(argv.LogFile)
			if !logFileExists {
				LogError("Logfile doesn't exists")
				return nil
			}
			config = &Config{
				Host:    argv.Host,
				LogFile: argv.LogFile,
				Token:   argv.Token,
				Note:    argv.Note,
			}
		} else {
			fileConfig := readConfig(configFile)
			logFile := fileConfig.LogFile
			host := fileConfig.Host
			token := fileConfig.Token
			note := fileConfig.Note
			if len(argv.LogFile) > 0 {
				logFile = argv.LogFile
			}
			logFileExists := validateLogFile(logFile)
			if !logFileExists {
				LogError("Logfile doesn't exists")
				return nil
			}
			if len(argv.Host) > 0 {
				host = argv.Host
			}
			if len(argv.Token) > 0 {
				token = argv.Token
			}
			if len(argv.Note) > 0 {
				note = argv.Note
			}
			config = &Config{
				Host:    host,
				LogFile: logFile,
				Token:   token,
				Note:    note,
			}
		}

		logFileExists := validateLogFile(config.LogFile)
		if !logFileExists {
			LogError("Logfile doesn't exists")
			return nil
		}

		if argv.UpdateEverything && !argv.DoUpdate {
			LogInfo("Ignoring -a! --update is not set! If you want to update everything, use -a and -u")
		}

		useLog := false

		//	iptablesparser.ParseFileByLines(config.LogFile, func(entry *iptablesparser.LogEntry) {
		//fmt.Println(entry)
		//})

		reportData := ReportStruct{
			Token:     config.Token,
			StartTime: time.Now().Unix(),
			IPs:       []IPData{},
		}
		if len(argv.CustomIPs) > 0 {
			ipsets := strings.Split(argv.CustomIPs, ";")
			for _, ipset := range ipsets {
				ipdat := strings.Split(ipset, ",")
				if len(ipdat) < 2 || len(ipdat) > 3 {
					LogInfo("Port missing for IP \"" + ipset + "\"! Skipping")
					continue
				}
				ip := ipdat[0]
				port, err := strconv.Atoi(ipdat[1])
				if err != nil {
					LogError("Port (" + ipdat[1] + ") no valid port!")
					continue
				}
				count := 1
				if len(ipdat) == 3 {
					count, err = strconv.Atoi(ipdat[2])
					if err != nil {
						LogError("Port (" + ipdat[2] + ") no valid count!")
						continue
					}
				}
				reportData.IPs = append(reportData.IPs, IPData{
					IP: ip,
					Ports: []IPPortReport{
						IPPortReport{
							Port:  port,
							Times: fillIntArray(count, 1),
						},
					},
				})
			}

			reportIPs(*config, reportData, argv.IgnoreCert)
		} else {
			fmt.Println("Currently not supportet!")
		}

		if useLog {
			runCommand(nil, "cat "+config.LogFile+" >> "+config.LogFile+"_1")
			runCommand(nil, "echo -n > "+config.LogFile)
		}

		if argv.DoUpdate {
			FetchIPs(config, configFile, argv.UpdateEverything, argv.IgnoreCert)
		}

		return nil
	},
}

func fillIntArray(size, value int) []int {
	arr := make([]int, size)
	for i := 0; i < size; i++ {
		arr[i] = value
	}
	return arr
}

func reportIPs(config Config, reportData ReportStruct, ignorecert bool) {
	if len(reportData.IPs) == 0 {
		LogInfo("Nothing to do")
		return
	}
	jsondata, err := json.Marshal(reportData)
	if err != nil {
		LogCritical("Error creating json:" + err.Error())
		return
	}
	res, err := request(config.Host, "reportnew", jsondata, ignorecert)
	if err != nil {
		LogCritical("Error doing rest call: " + err.Error())
		return
	}
	LogInfo("Server response: " + res)
}
