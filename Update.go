package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/mkideal/cli"
)

type updateConfT struct {
	cli.Helper
	Host       string `cli:"r,host" usage:"Specify the host to send the data to"`
	Token      string `cli:"t,token" usage:"Specify the token required by uploading hosts"`
	ConfigName string `cli:"C,config" usage:"Specify the config to use" dft:"config.json"`
	FetchAll   bool   `cli:"a,all" usage:"Fetches everything"`
	IgnoreCert bool   `cli:"i,ignorecert" usage:"Ignore invalid certs" dft:"false"`
}

var updateCMD = &cli.Command{
	Name:    "update",
	Aliases: []string{"u", "upd", "update"},
	Desc:    "Download and apply IP filter",
	Argv:    func() interface{} { return new(updateConfT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*updateConfT)
		if os.Getuid() != 0 {
			fmt.Println("You need to be root!")
			return nil
		}
		if !isIpsetInstalled(true) {
			return nil
		}

		blocklistName := getBlocklistName(argv.ConfigName)
		setupIPset(blocklistName)

		logStatus, configFile := createAndValidateConfigFile(argv.ConfigName)
		var config *Config
		if logStatus < 0 {
			return nil
		} else if logStatus == 0 {
			fmt.Println("Config empty. Using parameter as config. You can change them with <config>. Try 'triplink help config' for more information.")
			if len(argv.Host) == 0 || len(argv.Token) == 0 {
				fmt.Println("There is no such config file! You have to set all arguments. Try 'triplink help report'")
				return nil
			}
			config = &Config{
				Host:  argv.Host,
				Token: argv.Token,
			}
		} else {
			fileConfig := readConfig(configFile)
			if len(argv.Host) > 0 {
				fileConfig.Host = argv.Host
			}
			if len(argv.Token) > 0 {
				fileConfig.Token = argv.Token
			}
			config = fileConfig
		}

		err := FetchIPs(config, configFile, argv.FetchAll, argv.IgnoreCert, blocklistName)
		if err != nil {
			fmt.Println("Error fetching Update: " + err.Error())
		}

		return nil
	},
}

//FetchIPs fetches IPs and puts them into a blocklist
func FetchIPs(c *Config, configFile string, fetchAll, ignoreCert bool, blocklistName string) error {
	if c.Filter.Since == 0 {
		fetchAll = true
	}

	if fetchAll {
		c.Filter.Since = 0
	}
	requestData := FetchRequest{
		Token:  c.Token,
		Filter: c.Filter,
	}
	js, err := json.Marshal(requestData)
	if err != nil {
		return err
	}

	data, _, err := request(c.Host, "fetch", js, ignoreCert, true)
	data = strings.ReplaceAll(data, "\n", "")
	if err != nil || data == "[]" {
		if data == "\"[]\"" {
			LogInfo("Nothing to do (updating)")
		}
		return err
	}

	var fetchresponse FetchResponse
	err = json.Unmarshal([]byte(data), &fetchresponse)
	if err != nil {
		return err
	}

	c.Filter.Since = fetchresponse.CurrentTimestamp
	c.save(configFile)
	if fetchresponse.Full || fetchAll {
		flusIPset()
	}

	blockIPs(fetchresponse.IPs, blocklistName, c)
	backupIPs(configFile, true, false)
	return nil
}

func blockIPs(ips []IPList, blocklistName string, config *Config) {
	addCount := 0
	remCount := 0
	for _, ip := range ips {
		if ip.Deleted == 1 {
			if ipsetRemoveIP(ip.IP, blocklistName) {
				remCount++
			}
		} else {
			if ipsetAddIP(ip.IP, blocklistName) {
				addCount++
			}
		}
	}

	errorCreatingtriplinkChain := checkChain("triplink")
	if errorCreatingtriplinkChain {
		LogError("Couldn't create triplinkchain! Blocking might be unavailable")
		return
	}

	//check/create blocklistname-chain
	errorCreatingblnChain := checkChain(blocklistName)
	if errorCreatingblnChain {
		LogError("Couldn't create blocklist-chain! Blocking might be unavailable")
		return
	}
	blocklistTCPUDPname := blocklistName + "_tcp_udp"
	//check/create blocklistname_-chain
	errorCreatingblnportChain := checkChain(blocklistTCPUDPname)
	if errorCreatingblnportChain {
		LogError("Couldn't create blocklist_tcp_udp-chain! Blocking might be unavailable")
		return
	}

	commands := []iptableCommand{
		//INPUT -> triplink
		iptableCommand{
			"A",
			"INPUT -j triplink",
		},
		//triplink -> bloclist_config if not udp
		iptableCommand{
			"I",
			"triplink ! -p udp -j " + blocklistName,
		},
		//DROP if not tcp
		iptableCommand{
			"I",
			blocklistName + " ! -p tcp -m set --match-set " + blocklistName + " src -j DROP",
		},
		//triplink -> bloclist_config if not udp
		iptableCommand{
			"I",
			"triplink -j " + blocklistTCPUDPname,
		},
		//RETURN back to triplink
		iptableCommand{
			"A",
			blocklistName + " -j RETURN",
		},
		//DROP TCP PORTS
		iptableCommand{
			"I",
			blocklistTCPUDPname + " -p tcp -m set --match-set " + blocklistName + " src -m multiport --dports " + config.PortsToBlock + " -j DROP",
		},
		iptableCommand{
			"I",
			blocklistTCPUDPname + " -p udp -m set --match-set " + blocklistName + " src -m multiport --dports " + config.PortsToBlock + " -j DROP",
		},
		iptableCommand{
			"A",
			blocklistTCPUDPname + " -j RETURN",
		},
		iptableCommand{
			"A",
			"triplink -j RETURN",
		},
	}

	for _, cmd := range commands {
		if !runIptablesAction(cmd) {
			return
		}
	}
	if addCount > 0 || remCount > 0 {
		LogInfo("Successfully added " + strconv.Itoa(addCount) + " and removed " + strconv.Itoa(remCount) + " IPs")
	}

	//if activateIPset(blocklistName, config.PortsToBlock) {
	//}
}

/*
func activateIPset(blocklistName string, blockedPorts string) bool {
	if iptableHasRule(blocklistName, blockedPorts) {
		return true
	}
	_, err := runCommand(nil, "iptables -I INPUT 1 -m set --match-set "+blocklistName+" src -p tcp --match multiport --dports "+blockedPorts+" -j DROP")
	if err != nil {
		LogError("Couldn't activate iptable set. Blocking might be unavailable: " + err.Error() + " -> \"" + "iptables -I INPUT 1 -m set --match-set " + blocklistName + " src -p tcp --match multiport --dports " + blockedPorts + " -j DROP" + "\"")
		return false
	}
	return true
}
func iptableHasRule(blocklistName string, blockedPorts string) bool {
	_, err := runCommand(nil, "iptables -C INPUT -m set --match-set "+blocklistName+" src -p tcp --match multiport --dports "+blockedPorts+" -j DROP")
	return err == nil
}
*/

type iptableCommand struct {
	action, args string
}

func runIptablesAction(cmd iptableCommand, igncheck ...bool) bool {
	do := false
	if len(igncheck) == 0 || (len(igncheck) > 0 && !igncheck[0]) {
		_, err := runCommand(nil, "iptables -C "+cmd.args)
		if err != nil {
			do = true
		}
	} else {
		do = true
	}
	if do {
		_, err := runCommand(nil, "iptables -"+cmd.action+" "+cmd.args)
		fmt.Println("iptables -" + cmd.action + " " + cmd.args)
		if err != nil {
			LogError("Can't run \"iptables -" + cmd.action + " " + cmd.args + "\" " + err.Error())
			return false
		}
	}
	return true
}

func flusIPset() {
	runCommand(nil, "ipset flush blocklist")
}

func ipsetAddIP(ip string, blocklistName string) bool {
	valid, _ := isIPValid(ip)
	if valid {
		_, err := runCommand(nil, "ipset add "+blocklistName+" "+ip)
		return err == nil
	}
	return false
}

func ipsetRemoveIP(ip string, blocklistName string) bool {
	valid, _ := isIPValid(ip)
	if valid {
		_, err := runCommand(nil, "ipset del "+blocklistName+" "+ip)
		return err == nil
	}
	return false
}

func isIpsetInstalled(showerror bool) bool {
	_, err := runCommand(nil, "ipset help")
	if err != nil {
		if showerror {
			LogInfo("You need to install 'ipset' to run this command!")
		}
		return false
	}
	return true
}

func hasBlocklist(blocklistName string) bool {
	_, err := runCommand(nil, "ipset list "+blocklistName)
	return err == nil
}

func createBlocklist(blocklistName string) bool {
	_, err := runCommand(nil, "ipset create "+blocklistName+" nethash")
	return err == nil
}

func setupIPset(blocklistName string) {
	if !hasBlocklist(blocklistName) {
		if !createBlocklist(blocklistName) {
			LogError("Couldn't create blocklist! Exiting")
			os.Exit(1)
		}
	}
}

func getBlocklistName(configName string) string {
	if strings.Contains(configName, "/") {
		_, configName = path.Split(configName)
	}
	until := len(configName)
	if strings.Contains(configName, ".") {
		until = strings.LastIndex(configName, ".")
	}
	return "blocklist_" + configName[:until]
}

func checkChain(name string) bool {
	_, err := runCommand(nil, "iptables -L "+name)
	if err != nil {
		_, err = runCommand(nil, "iptables -N "+name)
		if err != nil {
			return true
		}
	}
	return false
}
