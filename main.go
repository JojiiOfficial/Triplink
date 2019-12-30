package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mkideal/cli"
)

var help = cli.HelpCommand("display help information")
var logPrefix = ""
var showTimeInLog = true
var version = "0.4"
var verboseLevel int

type argT struct {
	cli.Helper
	Version bool `cli:"v,version" usage:"Show version"`
}

var root = &cli.Command{
	Argv: func() interface{} { return new(argT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		if argv.Version {
			fmt.Println("Triplink V."+version, runtime.GOOS+"/"+runtime.GOARCH)
		} else {
			fmt.Println("Commands:\n\n" +
				"help           display help information\n" +
				"createConfig   Create a new configuration file (aliases cc,cconf,createconf,createconfig)\n" +
				"editConfig     Edit a configuration file (aliases econf,editconfig,ec,editc,edconf)\n" +
				"deleteConfig   Delete a configuration file (aliases delconf,deleteconfig,dc,dconf,delc)\n" +
				"viewConfig     View a configuration file (aliases vconf,vc,viewc,showconf,showconfig,config,conf,confshow,confview)\n" +
				"install        Install a cronjob easily to automate the updating/reporting process\n" +
				"update         Download and apply IP filter(aliases u,upd,update)\n" +
				"rules          change iptable rules created by triplink(aliases rule,rul)\n" +
				"ipinfo         Show info for an IP (aliases info,showip,ipdata,ipd,ii)\n" +
				"restore        restore ipset and iptables (aliases res,restore,rest)\n" +
				"backup         backups ipset(-s) and (iptables with -t) (aliases b,bak,backup)\n" +
				"deletebackup   delete backups from ipset(-s) and (iptables -t) (aliases db,delbak,delbackup,deleteb,dback,delb)\n" +
				"report         Reports all changes (aliases r,report,repo)")
		}
		return nil
	},
}

func main() {
	if err := cli.Root(root,
		cli.Tree(help),
		cli.Tree(createConfCMD),
		cli.Tree(editConfCMD),
		cli.Tree(deleteConfCMD),
		cli.Tree(viewConfCMD),
		cli.Tree(installCMD),
		cli.Tree(updateCMD),
		cli.Tree(rulesCMD),
		cli.Tree(restoreCMD),
		cli.Tree(backupCMD),
		cli.Tree(delBackupCMD),
		cli.Tree(ipinfoCMD),
		cli.Tree(reportCMD),
		cli.Tree(pingCMD),
	).Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
