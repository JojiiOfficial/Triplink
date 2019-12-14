package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	iptablesparser "github.com/JojiiOfficial/Iptables-log-parser"

	"github.com/mkideal/cli"
)

type reportT struct {
	cli.Helper
	LogFile          string `cli:"f,file" usage:"Specify the file to read the logs from"`
	Host             string `cli:"r,host" usage:"Specify the host to send the data to"`
	Token            string `cli:"t,token" usage:"Specify the token required by uploading hosts"`
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

		logStatus, configFile := createAndValidateConfigFile(argv.ConfigName)
		var config *Config
		if logStatus < 0 {
			return nil
		} else if logStatus == 0 {
			fmt.Println("Config empty. Using parameter as config. You can change them with <config>. Try 'triplink help config' for more information.")
			if len(argv.Host) == 0 || len(argv.LogFile) == 0 || len(argv.Token) == 0 {
				fmt.Println("There is no such config file! You have to set all arguments. Try 'triplink help report'")
				return nil
			}
			config = &Config{
				Host:    argv.Host,
				LogFile: argv.LogFile,
				Token:   argv.Token,
			}
		} else {
			fileConfig := readConfig(configFile)
			logFile := fileConfig.LogFile
			host := fileConfig.Host
			token := fileConfig.Token
			if len(argv.LogFile) > 0 {
				logFile = argv.LogFile
			}
			if len(argv.Host) > 0 {
				host = argv.Host
			}
			if len(argv.Token) > 0 {
				token = argv.Token
			}
			config = &Config{
				Host:    host,
				LogFile: logFile,
				Token:   token,
				Filter:  fileConfig.Filter,
			}
		}

		logFileExists := validateLogFile(config.LogFile)
		if !logFileExists && len(argv.CustomIPs) == 0 {
			LogError("The value for \"logfile\" is missing in the config file!")
			return nil
		}

		if argv.UpdateEverything && !argv.DoUpdate {
			LogInfo("Ignoring -a! --update is not set! If you want to update everything, use -a and -u")
		}

		useLog := len(argv.CustomIPs) == 0

		reportData := ReportStruct{
			Token:     config.Token,
			StartTime: time.Now().Unix(),
			IPs:       []IPData{},
		}

		responseSuccess := false

		if !useLog {
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

			responseSuccess = reportIPs(*config, reportData, argv.IgnoreCert)
		} else {
			startTime := int64(-1)
			ipMap := make(map[string][]IPTimePort)
			err := iptablesparser.ParseFileByLines(config.LogFile, func(log *iptablesparser.LogEntry) {
				ip := log.Src
				if startTime == -1 {
					startTime = log.Time.Unix()
				}

				timeDiff := (int)(log.Time.Unix() - startTime)
				ipTimePortToAdd := IPTimePort{
					Port: log.DestPort,
					Time: timeDiff,
				}

				if _, ok := ipMap[ip]; !ok {
					ipMap[ip] = []IPTimePort{ipTimePortToAdd}
				} else {
					ipMap[ip] = append(ipMap[ip], ipTimePortToAdd)
				}
			})
			if err != nil {
				LogCritical("Couldn't parse logfile: " + err.Error())
				return nil
			}
			var ipdata []IPData
			reportData.StartTime = startTime
			for ip, timesWithPorts := range ipMap {
				portTimeMap := make(map[int][]int)
				for _, tp := range timesWithPorts {
					if _, ok := portTimeMap[tp.Port]; !ok {
						portTimeMap[tp.Port] = []int{tp.Time}
					} else {
						portTimeMap[tp.Port] = append(portTimeMap[tp.Port], tp.Time)
					}
				}
				ipports := []IPPortReport{}
				for port, times := range portTimeMap {
					ipports = append(ipports, IPPortReport{
						Port:  port,
						Times: times,
					})
				}
				ipdata = append(ipdata, IPData{
					IP:    ip,
					Ports: ipports,
				})
			}
			reportData.IPs = ipdata
			responseSuccess = reportIPs(*config, reportData, argv.IgnoreCert)
		}

		if useLog && responseSuccess {
			runCommand(nil, "cat "+config.LogFile+" >> "+config.LogFile+"_1")
			runCommand(nil, "echo -n > "+config.LogFile)
		} else if !responseSuccess && useLog {
			LogInfo("Keeping logs until report was successful")
			return nil
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

func reportIPs(config Config, reportData ReportStruct, ignorecert bool) bool {
	if len(reportData.IPs) == 0 {
		LogInfo("Nothing to do")
		return true
	}
	jsondata, err := json.Marshal(reportData)
	if err != nil {
		LogCritical("Error creating json:" + err.Error())
		return false
	}
	res, isStatus, err := request(config.Host, "reportnew", jsondata, ignorecert, true)
	if err != nil {
		LogCritical("Error doing rest call: " + err.Error())
		return false
	}
	if isStatus {
		status, _ := responseToStatus(res)
		LogInfo("Server response: " + status.StatusMessage)
		return (status.StatusMessage == "success")
	}
	LogError("Got weird server response: " + res)
	return false
}
