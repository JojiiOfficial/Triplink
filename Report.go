package main

import (
	"encoding/json"
	"fmt"
	"os"
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
	Note             string `cli:"n,note" usage:"Sends a very short description"`
	DoUpdate         bool   `cli:"u,update" usage:"Specify if the client should update after the report" dft:"false"`
	UpdateEverything bool   `cli:"a,all" usage:"Specify if the client should update everything if update is set" dft:"false"`
	CustomIPs        string `cli:"c,custom" usage:"Report a custom IPset"`
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

		ipsToReport := []IPset{}
		useLog := len(argv.CustomIPs) == 0
		if !useLog {
			LogInfo("using arguments")
			ips := strings.Split(strings.Trim(argv.CustomIPs, " "), ";")
			for _, ip := range ips {
				ip = strings.Trim(ip, " ")
				iptrp := ""
				reason := 1
				val := 0
				if strings.Contains(ip, ",") {
					dat := strings.Split(ip, ",")
					dat[0] = strings.Trim(dat[0], " ")
					iReason, err := strconv.Atoi(strings.Trim(dat[1], " "))
					if err == nil {
						reason = iReason
					}
					if len(dat) == 3 {
						ival, err := strconv.Atoi(strings.Trim(dat[2], " "))
						if err == nil {
							val = ival
						}
					}
					iptrp = strings.Trim(dat[0], " ")
				} else {
					iptrp = ip
				}
				ipsToCheck := []string{}
				if strings.Contains(iptrp, "/") {
					cidr, err := strconv.Atoi(strings.Split(iptrp, "/")[1])
					if err != nil {
						LogError("CIDR is no int! Skipping " + iptrp)
						continue
					}
					if cidr < 20 {
						LogError("You really want to report more than 256 IPs? I don't think so")
						continue
					}
					iplist, err := cidrToIPlist(iptrp)
					if err != nil {
						LogError("Error parsing CIDR:" + err.Error())
						LogInfo("Skipping CIDR range!")
						continue
					}
					for _, cip := range iplist {
						ipsToCheck = append(ipsToCheck, cip)
					}
				} else {
					ipsToCheck = append(ipsToCheck, iptrp)
				}
				for _, icp := range ipsToCheck {
					valid, nvReason := isIPValid(icp)
					if valid {
						ipsToReport = append(ipsToReport, IPset{IP: icp, Reason: reason, Valid: val})
					} else {
						LogError("Ip is not valid: " + icp + " " + ipErrToString(nvReason) + " skipping")
					}
				}
			}
		} else {
			LogInfo("using log: " + config.LogFile)
			ipTime := make(map[string]([]time.Time))
			err := iptablesparser.ParseFileByLines(config.LogFile, func(log *iptablesparser.LogEntry) {
				_, has := ipTime[log.Src]
				if has {
					ipTime[log.Src] = append(ipTime[log.Src], log.Time)
				} else {
					ipTime[log.Src] = append([]time.Time{}, log.Time)
				}
			})

			if err != nil {
				LogError("Can't read File: " + err.Error())
			}
			for ip, t := range ipTime {
				valid, _ := isIPValid(ip)
				if !valid {
					continue
				}
				reason := IPrequestTimesToReason(t)
				ipsToReport = append(ipsToReport, IPset{ip, reason, 0})
			}
		}

		if len(ipsToReport) > 0 {
			reportStruct := ReportIPStruct{Token: config.Token, Ips: ipsToReport, Note: config.Note}
			js, err := json.Marshal(reportStruct)
			if err != nil {
				panic(err)
			}

			resp, err := request(config.Host, "report", js, argv.IgnoreCert)
			if err != nil {
				LogCritical("error making request: " + err.Error())
			} else {
				LogInfo(resp)
			}

		} else {
			LogInfo("Nothing to do (reporting)")
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

const maxCountToBrute int = 20

//IPrequestTimesToReason returns a reason based on the frequency of connect attempts
func IPrequestTimesToReason(timeList []time.Time) int {
	if len(timeList) == 0 {
		return -1
	} else if len(timeList) <= 2 {
		//return Scanner
		return 1
	}

	bruteToleranceLine := (int)(len(timeList) / 9)

	lastPing := timeList[0]
	spamCounter := 0
	scanCounter := 0
	bruteRow := 0
	bruteTolerance := 0
	for _, t := range timeList {
		diff := t.Sub(lastPing).Minutes()
		if diff < 1 {
			bruteRow++
		} else if diff <= 10 {
			spamCounter++
			if bruteRow < maxCountToBrute {
				if bruteTolerance > bruteToleranceLine {
					bruteRow = 0
				} else {
					bruteTolerance++
				}
			} else {
				return 3
			}
		} else {
			scanCounter++
			if bruteRow < maxCountToBrute {
				if bruteTolerance > bruteToleranceLine {
					bruteRow = 0
				} else {
					bruteTolerance++
				}
			}
		}
		lastPing = t
	}

	if bruteRow >= 14 {
		return 3
	}

	if spamCounter >= 7 {
		return 2
	}

	a, b := percentRelation(scanCounter, spamCounter)
	if a > (b*1.85) && scanCounter < 15 {
		return 1
	}
	return 2
}

func percentRelation(a, b int) (float32, float32) {
	ges := a + b
	if a+b == 0 {
		return 0, 0
	}
	return (float32)(a * 100 / ges), (float32)(b * 100 / ges)
}

func avgTimeDiff(timeList []time.Time) float32 {
	deltaTime := float32(0)
	lastPing := timeList[0]
	for _, t := range timeList {
		deltaTime += float32(t.Sub(lastPing).Minutes())
	}
	return ((float32)(deltaTime) / float32(len(timeList)))
}
