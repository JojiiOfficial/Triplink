package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strings"
)

//Config the global config struct
type Config struct {
	Host          string      `json:"host"`
	LogFile       string      `json:"logfile"`
	Token         string      `json:"token"`
	Filter        FetchFilter `json:"fetchFilter"`
	Note          string      `json:"note"`
	ShowTimeInLog bool        `json:"showLogTime"`
}

func getConfPath(homeDir string) string {
	return homeDir + "/" + ".tripwirereporter/"
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
