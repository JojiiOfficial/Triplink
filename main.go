package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mkideal/cli"
)

var logPrefix = ""
var showTimeInLog = true
var version = "0.6"
var verboseLevel int

func main() {
	rootCommand := os.Args[1:]

	err := cli.Root(root,
		cli.Tree(rootConfigCMD,
			cli.Tree(cli.HelpCommand(displayHelpInformation)),
			cli.Tree(createConfCMD),
			cli.Tree(editConfCMD),
			cli.Tree(viewConfCMD),
			cli.Tree(deleteConfCMD),
		),
		cli.Tree(rootBackupCMD,
			cli.Tree(cli.HelpCommand(displayHelpInformation)),
			cli.Tree(backupCMD),
			cli.Tree(restoreCMD),
			cli.Tree(delBackupCMD),
		),
		cli.Tree(cli.HelpCommand(displayHelpInformation)),
		cli.Tree(fetchCMD),
		cli.Tree(installCMD),
		cli.Tree(rulesCMD),
		cli.Tree(pingCMD),
		cli.Tree(ipinfoCMD),
		cli.Tree(reportCMD),
	).Run(rootCommand)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return
}

type argHelperT struct {
	cli.Helper
}

type argT struct {
	cli.Helper
	Version bool `cli:"v,version" usage:"Show version"`
}

var rootConfigCMD = &cli.Command{
	Name: "config",
	Desc: "Create, view, delete or modify a configuration file",
	Argv: func() interface{} { return new(argHelperT) },
	Fn: func(ctx *cli.Context) error {
		fmt.Println("Commands:\n\n" +
			"  help     Display help information\n" +
			"  create   Create a new configuration file (aliases cc,cconf,createconf,createconfig)\n" +
			"  edit     Edit a configuration file (aliases econf,editconfig,ec,editc,edconf)\n" +
			"  delete   Delete a configuration file (aliases delconf,deleteconfig,dc,dconf,delc)\n" +
			"  view     View a configuration file (aliases vconf,vc,viewc,showconf,showconfig,config,conf,confshow,confview)")
		return nil
	},
}

var rootBackupCMD = &cli.Command{
	Name: "backup",
	Desc: "Create, delete or restore iptables+ipset backups",
	Argv: func() interface{} { return new(argHelperT) },
	Fn: func(ctx *cli.Context) error {
		fmt.Println("Commands:\n\n" +
			"  help           Display help information\n" +
			"  create         backups ipset(-s) and (iptables with -t) (aliases b,bak,backup)\n" +
			"  restore        restore ipset and iptables (aliases res,restore,rest)\n" +
			"  deletebackup   delete backups from ipset(-s) and (iptables -t) (aliases db,delbak,delbackup,deleteb,dback,delb)")
		return nil
	},
}

var root = &cli.Command{
	Argv: func() interface{} { return new(argT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		if argv.Version {
			fmt.Println("Triplink V."+version, runtime.GOOS+"/"+runtime.GOARCH)
		} else {
			fmt.Println("Commands:\n\n" +
				"  config    Create, view, delete or modify a configuration file\n" +
				"  backup    Create, delete or restore iptables+ipset backups\n" +
				"  fetch     Fetch and block IPs matching the filter assigned to the token(aliases u,upd,update)\n" +
				"  install   Setup automatic reports/updates easily(aliases install)\n" +
				"  rules     change iptable rules created by triplink(aliases rule,rul)\n" +
				"  ping      Checks connection to server is successful(aliases p,ping,pin,pi)\n" +
				"  ipinfo    Show info for an IP(aliases info,showip,ipdata,ipd,ii)\n" +
				"  report    Reports all changes(aliases r,report,repo)")
		}
		return nil
	},
}
