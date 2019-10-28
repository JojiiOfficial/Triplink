package main

import (
	"encoding/json"
	"fmt"
	"os"
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
		logStatus, configFile := createAndValidateConfigFile()
		var config *Config
		if logStatus < 0 {
			return nil
		} else if logStatus == 0 {
			fmt.Println("Config empty. Using parameter as config. You can change them with <config>. Try 'twreporter help config' for more information.")
			if len(argv.Host) == 0 || len(argv.LogFile) == 0 || len(argv.Token) == 0 {
				fmt.Println("There is no config file! You have to set all arguments. Try 'twreporter help report'")
				return nil
			}
			logFileExists := validateLogFile(argv.LogFile)
			if !logFileExists {
				fmt.Println("Logfile doesn't exists")
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
			logFileExists := validateLogFile(logFile)
			if !logFileExists {
				fmt.Println("Logfile doesn't exists")
				return nil
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
			}
		}

		logFileExists := validateLogFile(config.LogFile)
		if !logFileExists {
			fmt.Println("Logfile doesn't exists")
			return nil
		}

		if argv.UpdateEverything && !argv.DoUpdate {
			fmt.Println("Ignoring -a! --update is not set! If you want to update everything, use -a and -u")
		}

		fmt.Println("using log: ", config.LogFile)
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
			fmt.Println("Can't read File: ", err.Error())
		}

		ipsToReport := []IPset{}
		for ip, t := range ipTime {
			valid, _ := isIPValid(ip)
			if !valid {
				continue
			}
			reason := IPrequestTimesToReason(t)
			ipsToReport = append(ipsToReport, IPset{ip, reason})
		}

		if len(ipsToReport) > 0 {

			reportStruct := ReportIPStruct{Token: config.Token, Ips: ipsToReport}

			js, err := json.Marshal(reportStruct)
			if err != nil {
				panic(err)
			}

			resp, err := request(config.Host+"/report", js)
			if err != nil {
				fmt.Println("error making request: " + err.Error())
			} else {
				fmt.Print(resp)
			}

		} else {
			fmt.Println("Nothing to do (reporting)")
		}
		runCommand(nil, "cat "+config.LogFile+" >> "+config.LogFile+"_1")
		runCommand(nil, "echo -n > "+config.LogFile)

		if argv.DoUpdate {
			FetchIPs(config, configFile, argv.UpdateEverything)
		}

		return nil
	},
}

const maxCountToBrute int = 20

//IPrequestTimesToReason returns a reason based on the frequency of connect attempts
func IPrequestTimesToReason(timeList []time.Time) int {
	if len(timeList) == 0 {
		return -1
	} else if len(timeList) <= 6 {
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

	if spamCounter > 10 {
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
