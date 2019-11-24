package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mkideal/cli"
)

type ipinfoT struct {
	cli.Helper
	IPs               []string `cli:"*i,ips" usage:"Show info for these IPs"`
	ConfigName        string   `cli:"C,config" usage:"Specify the config to use" dft:"config.json"`
	LogFile           string   `cli:"f,file" usage:"Specify the file to read the logs from"`
	Host              string   `cli:"r,host" usage:"Specify the host to send the data to"`
	Token             string   `cli:"t,token" usage:"Specify the token required by uploading hosts"`
	Note              string   `cli:"n,note" usage:"Sends a very short description"`
	HideNoReportFound bool     `cli:"H,hide" usage:"Hide the message for an IP without reports"`
	IgnoreCert        bool     `cli:"ignorecert" usage:"Ignore invalid certs" dft:"false"`
}

var ipinfoCMD = &cli.Command{
	Name:    "ipinfo",
	Aliases: []string{"info", "showip", "ipdata", "ipd", "ii"},
	Desc:    "Show info for an IP",
	Argv:    func() interface{} { return new(ipinfoT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*ipinfoT)

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

		InitArrayParam(&argv.IPs, ",")
		ips := []string{}
		for _, ip := range argv.IPs {
			if v, _ := isIPValid(ip); !v {
				fmt.Println("IP \"" + ip + "\" isn't valid!")
				continue
			}
			ips = append(ips, ip)
		}
		if len(ips) == 0 {
			return errors.New("No valid IP found")
		}

		requ := IPInfoRequest{
			Token: config.Token,
			IPs:   ips,
		}

		d, err := json.Marshal(requ)
		if err != nil {
			return errors.New("Couldn't create json")
		}
		res, err := request(config.Host, "ipinfo", d, argv.IgnoreCert)
		if err != nil {
			return errors.New("Error doing rest call: " + err.Error())
		}
		res = strings.Trim(strings.ReplaceAll(res, "\n", ""), " ")
		if res == "-1" {
			return errors.New("Invalid token")
		} else if res == "2" {
			return errors.New("Server error")
		}
		var ipdata []IPInfoData
		err = json.Unmarshal([]byte(res), &ipdata)
		if err != nil {
			return errors.New("Error parsing response: " + err.Error())
		}

		displayIPdata(&ipdata, argv.HideNoReportFound)

		return nil
	},
}

func displayIPdata(ipdata *[]IPInfoData, hideNotFound bool) {
	for i, info := range *ipdata {
		var add string
		if i > 0 {
			add = "\n"
		}
		if len(info.Reports) > 0 {
			var max int
			for _, ce := range info.Reports {
				max += ce.Count
			}
			fmt.Println(add + "IP: " + info.IP + " (" + strconv.Itoa(max) + "x)")
			for _, report := range info.Reports {
				fmt.Println("  ", parseTime(report.Time), report.ReporterName, ":"+strconv.Itoa(report.Port), "("+strconv.Itoa(report.Count)+"x)")
			}
		} else if !hideNotFound {
			fmt.Println(add + "No report for " + info.IP)
		}
	}
}

func parseTime(unix int64) string {
	return time.Unix(unix, 0).Format(time.Stamp)
}

//InitArrayParam split parameter values
func InitArrayParam(sl *[]string, seperator string) {
	if len(*sl) == 0 {
		*sl = nil
		return
	}
	var e []string
	for _, hn := range *sl {
		if strings.Contains(hn, seperator) {
			for _, hh := range strings.Split(hn, seperator) {
				if len(hh) == 0 {
					continue
				}
				e = append(e, hh)
			}
		} else {
			if len(hn) == 0 {
				continue
			}
			e = append(e, hn)
		}
	}
	*sl = e
}
