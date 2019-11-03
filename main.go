package main

import (
	"fmt"
	"os"

	"github.com/mkideal/cli"
)

var help = cli.HelpCommand("display help information")
var logPrefix = ""
var showTimeInLog = true

type argT struct {
	cli.Helper
}

var root = &cli.Command{
	Desc: "this is root command",
	Argv: func() interface{} { return new(argT) },
	Fn: func(ctx *cli.Context) error {
		fmt.Println("Usage: twreporter <install/report/update/(view/create/delete)config/backup/restore> [-f,-r,-t,-o,-u,-a]")
		return nil
	},
}

func main() {
	if err := cli.Root(root,
		cli.Tree(help),
		cli.Tree(reportCMD),
		cli.Tree(createConfCMD),
		cli.Tree(deleteConfCMD),
		cli.Tree(viewConfCMD),
		cli.Tree(updateCMD),
		cli.Tree(installCMD),
		cli.Tree(backupCMD),
		cli.Tree(restoreCMD),
		cli.Tree(delBackupCMD),
		cli.Tree(editConfCMD),
	).Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
