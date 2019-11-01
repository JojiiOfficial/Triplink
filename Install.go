package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mkideal/cli"
)

type installT struct {
	cli.Helper
	ConfigName string `cli:"C,config" usage:"Specify the config to use" dft:"config.json"`
}

var installCMD = &cli.Command{
	Name:    "install",
	Aliases: []string{"install"},
	Desc:    "Setup automatic reports/updates easily",
	Argv:    func() interface{} { return new(installT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*installT)
		_ = argv

		reader := bufio.NewReader(os.Stdin)
		i, text := waitForMessage("What kind of system do you want to setup?\n[t] Tripwire\n[i] iptables/set\n> ", reader)
		if i == -1 {
			return nil
		}
		if i == 1 {
			text = strings.ToLower(text)
			if text == "t" {
				setTripwire(reader, argv.ConfigName)
			} else if text == "i" {
				setIP(reader, argv.ConfigName)
			} else {
				fmt.Println("What? Didn't understand '" + text + "'. Type 't' or 'i'")
				return nil
			}
		} else {
			return nil
		}

		//Tripwire related options
		//1. Update iplist (only blocking) periodically
		//2. Reporting + blocking periodically
		//3. Only reporting

		//firewall related options
		//1. Backup iptables & ipset
		//2. Restore iptables & ipset
		return nil
	},
}

func setIP(reader *bufio.Reader, config string) {
	i, opt := waitForMessage("Backup or Restore?\n[b] Backup\n"+
		"[r] Restore\n> ", reader)
	if i != 1 {
		return
	}
	opt = strings.ToLower(opt)
	mode := ""
	sMode := ""
	if opt == "b" {
		sMode = "Backup"
		mode = "b"
	} else if opt == "r" {
		sMode = "Restore"
		mode = "r"
	} else {
		fmt.Println("What? Didn't understand '" + opt + "'. Type 't' or 'i'")
		return
	}

	i, opt = waitForMessage("What to "+sMode+"?\n[1] IPset\n"+
		"[2] IPtables\n"+
		"[3] both\n> ", reader)
	if i != 1 {
		return
	}
}

func setTripwire(reader *bufio.Reader, config string) {
	config = getConfigPathFromHome(config)
	if handleConfig(config) {
		return
	}
	i, opt := waitForMessage("How should Tripwire act?\n[1] Fetch and block IPs from server based on a filter\n"+
		"[2] Report IPs based on a filter defined by you\n"+
		"[3] Report IPs only without blocking them\n> ", reader)
	if i != 1 {
		return
	}
	if opt != "1" && opt != "2" && opt != "3" {
		fmt.Println("What? Enter 1,2 or 3")
		return
	}

	i, text := waitForMessage("Do you want to update the filter assigned to this config [y/n] ", reader)
	if i == 1 && (text == "y" || text == "yes") {
		createFilter(config)
	}

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	i, text = waitForMessage("In which period do you want to run this action [min]: ", reader)
	if i != 1 {
		fmt.Println("Abort")
		return
	}
	in, err := strconv.Atoi(text)
	if err != nil {
		fmt.Println("Not an integer")
		return
	}
	if in < 0 || in > 69 {
		fmt.Println("Your range must be between 0 and 60")
		return
	}
	addCMD := ""
	if opt == "1" {
		addCMD = "u"
	} else if opt == "2" {
		addCMD = "r"
	} else if opt == "3" {
		addCMD = "r -u"
	} else {
		return
	}
	err = writeCrontab("*/" + text + " * * * * " + ex + " " + addCMD)
	if err != nil {
		fmt.Println("Error writing crontab: " + err.Error())
	} else {
		fmt.Println("Installed successfully")
	}
}

func writeCrontab(cronCommand string) error {
	f, err := os.OpenFile("/var/spool/cron/crontabs/root", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	f.WriteString(cronCommand + "\n")
	f.Close()
	return nil
}

func handleConfig(config string) bool {
	_, err := os.Stat(config)
	if err != nil {
		fmt.Println("Config not found. Create one with 'twreporter cc'.")
		return true
	}
	return false
}

func waitForMessage(question string, reader *bufio.Reader) (int, string) {
	fmt.Print(question)
	text, _ := reader.ReadString('\n')
	text = strings.ReplaceAll(text, "\n", "")
	if strings.ToLower(text) == "a" {
		return -1, ""
	}
	if len(text) > 0 {
		return 1, text
	}
	return 0, text
}
