package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/mkideal/cli"
)

type rulesT struct {
	cli.Helper
	ConfigName string `cli:"C,config" usage:"Specify the config to use" dft:"config.json"`
	Delete     bool   `cli:"d,delete" usage:"Delete iptable rules for given config. Disabe blocking IPs for given config"`
	Create     bool   `cli:"c,create" usage:"Create iptable rules for given config. Enable blocking IPs for given config"`
	Update     bool   `cli:"u,update" usage:"Update iptables rules for given config"`
	Yes        bool   `cli:"y,yes" usage:"Don't confirm deletion" dft:"false"`
	Verbose    int    `cli:"v,verbose" usage:"Specify how much logs should be displayed" dft:"0"`
}

func (argv *rulesT) Validate(ctx *cli.Context) error {
	if !checkBoolArrUnique(argv.Create, argv.Delete, argv.Update) {
		return errors.New("you need to enter a command")
	}
	return nil
}

func checkBoolArrUnique(bools ...bool) bool {
	hasTrue := false
	trueCount := 0
	for _, b := range bools {
		if b {
			hasTrue = true
			trueCount++
		}
	}
	return hasTrue && trueCount == 1
}

var rulesCMD = &cli.Command{
	Name:    "rules",
	Aliases: []string{"rule", "rul"},
	Desc:    "change iptable rules created by triplink",
	Argv:    func() interface{} { return new(rulesT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*rulesT)
		verboseLevel = argv.Verbose
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Couldn't retrieve homeDir!")
			return nil
		}
		if os.Getuid() != 0 {
			LogError("Only root can do this!")
			return nil
		}

		confPath := getConfPath(homeDir)
		confFile := getConfFile(confPath, argv.ConfigName)
		_, err = os.Stat(confFile)
		if err == nil {
			realConf := readConfig(confFile)

			//Warn/confirm if auto rules are disabled
			if !realConf.AutocreateIptables && !argv.Yes && (argv.Create || argv.Update) {
				cont, i := confirmInput("Auto iptable rules are disabled for this configuration. Do you want to continue anyway? [y/n] > ", bufio.NewReader(os.Stdin))
				if i != 1 || !cont {
					return nil
				}
			}

			success := false
			bln := getBlocklistName(confFile)
			if argv.Delete {
				deleteBlocklistIptableRules(bln)
				success = true
			} else if argv.Create {
				success = createIPtableRules(bln, realConf)
			} else if argv.Update {
				deleteBlocklistIptableRules(bln)
				success = createIPtableRules(bln, realConf)
			}
			if success {
				fmt.Println("Success")
			}
		} else {
			fmt.Println("Config not found. Nothing to do.")
		}
		return nil
	},
}

func createIPtableRules(blocklistName string, config *Config) bool {
	if !isIpsetInstalled(true) {
		return false
	}
	setupIPset(blocklistName)
	errorCreatingtriplinkChain := checkChain("triplink")
	if errorCreatingtriplinkChain {
		LogError("Couldn't create triplinkchain! Blocking might be unavailable")
		return false
	}

	//check/create blocklistname-chain
	errorCreatingblnChain := checkChain(blocklistName)
	if errorCreatingblnChain {
		LogError("Couldn't create blocklist-chain! Blocking might be unavailable")
		return false
	}
	blocklistTCPUDPname := blocklistName + "_tcp_udp"

	//check/create blocklistname_-chain
	errorCreatingblnportChain := checkChain(blocklistTCPUDPname)
	if errorCreatingblnportChain {
		LogError("Couldn't create blocklist_tcp_udp-chain! Blocking might be unavailable")
		return false
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
			return false
		}
	}
	return true
}

func deleteBlocklistIptableRules(blocklistName string) bool {
	blocklistTCPUDPname := blocklistName + "_tcp_udp"
	commandso := []iptableCommand{
		//remove triplink -> bloclist_config
		iptableCommand{
			"D",
			"triplink ! -p udp -j " + blocklistName,
		},
		//remove triplink -> bloclist_config_tcp_udp
		iptableCommand{
			"D",
			"triplink -j " + blocklistTCPUDPname,
		},
		//Flush blocklist_config
		iptableCommand{
			"F",
			blocklistName,
		},
		//Flush blocklist_config_tcp_udp
		iptableCommand{
			"F",
			blocklistTCPUDPname,
		},
		//Flush blocklist_config
		iptableCommand{
			"X",
			blocklistName,
		},
		//Flush blocklist_config_tcp_udp
		iptableCommand{
			"X",
			blocklistTCPUDPname,
		},
	}

	for _, cmd := range commandso {
		if !runIptablesAction(cmd, true) {
			continue
		}
	}
	runCommand(nil, "iptables -X "+blocklistName)
	return true
}
