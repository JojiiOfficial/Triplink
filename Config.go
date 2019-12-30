package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strconv"
	"strings"
)

//Config the global config struct
type Config struct {
	Host               string      `json:"host"`
	LogFile            string      `json:"logfile"`
	Token              string      `json:"token"`
	Filter             FetchFilter `json:"fetchFilter"`
	ShowTimeInLog      bool        `json:"showLogTime"`
	PortsToBlock       string      `json:"portsToBlock"`
	AutocreateIptables bool        `json:"createIPtableRules"`
}

func getConfPath(homeDir string) string {
	return homeDir + "/" + ".triplink/"
}

func getConfFile(confPath, confName string) string {
	if !strings.HasSuffix(confName, ".json") {
		confName += ".json"
	}
	return confPath + confName
}

func getConfigPathFromHome(confName string) string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return getConfFile(getConfPath(usr.HomeDir), confName)
}

func readConfig(file string) *Config {
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	res := Config{}
	err = json.Unmarshal(dat, &res)
	if err != nil {
		panic(err)
	}
	return &res
}

func (config *Config) save(configFile string) error {
	sConf, err := json.Marshal(config)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(configFile, []byte(string(sConf)), 0600)
	if err != nil {
		return err
	}
	return nil
}

func validateLogFile(logfile string) bool {
	_, err := os.Stat(logfile)
	if err != nil {
		return false
	}
	return true
}

func createAndValidateConfigFile(confName string) (int, string) {
	if len(confName) == 0 {
		LogError("No config name given")
		os.Exit(1)
		return -1, ""
	}
	if !strings.HasSuffix(confName, ".json") {
		confName += ".json"
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Couldn't retrieve homeDir!")
		return -1, ""
	}
	confPath := getConfPath(homeDir)
	confFile := getConfFile(confPath, confName)
	_, err = os.Stat(confPath)
	if err != nil {
		err = os.MkdirAll(confPath, os.ModePerm)
		if err != nil {
			fmt.Println("Couldn't create configpath")
			return -1, ""
		}
		_, err = os.Create(confFile)
		if err != nil {
			fmt.Println("Couldn't create configfile")
			return -1, ""
		}
	}
	confStat, err := os.Stat(confFile)
	if err != nil {
		_, err = os.Create(confFile)
		if err != nil {
			fmt.Println("Couldn't create configfile")
			return -1, ""
		}
	}
	confStat, err = os.Stat(confFile)
	if err != nil {
		fmt.Println("Couldn't create configfile")
		return -1, ""
	}
	if confStat.Size() == 0 {
		return 0, confFile
	}

	return 1, confFile
}

func getHome() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Couldn't retrieve homeDir!")
		os.Exit(1)
		return ""
	}
	return homeDir
}

func validatePortsParam(param string) (string, error) {
	if strings.Contains(param, ",") {
		ports := strings.Split(param, ",")
		for _, port := range ports {
			if !isSinglePortParamValid(port) {
				return "", errors.New("check your ports")
			}
		}
	} else {
		if !isSinglePortParamValid(param) {
			return "", errors.New("port must be an integer. Ranges are defined using a '-'. For example 100-200")
		}
	}
	return strings.ReplaceAll(param, "-", ":"), nil
}

func isSinglePortParamValid(portParam string) bool {
	if strings.Contains(portParam, "-") {
		return isPortRangeValid(portParam)
	}
	return isPortValid(portParam)
}

func isPortRangeValid(portrange string) bool {
	ports := strings.Split(portrange, "-")
	if len(ports) == 2 {
		var start, end int
		if isPortValid(ports[0]) && isPortValid(ports[1]) {
			start, _ = strconv.Atoi(ports[0])
			end, _ = strconv.Atoi(ports[1])
			if end > start {
				return true
			}
		}
	}
	return false
}

func isPortValid(port string) bool {
	i, err := strconv.Atoi(port)
	if err != nil || (i < 0 || i > 65535) {
		return false
	}
	return true
}
