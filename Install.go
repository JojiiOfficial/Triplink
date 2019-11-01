package main

import (
	"github.com/mkideal/cli"
)

type installT struct {
	cli.Helper
	ConfigName string `cli:"C,config" usage:"Secify the config to use" dft:"config.json"`
}

var installCMD = &cli.Command{
	Name:    "install",
	Aliases: []string{"install"},
	Desc:    "Setup automatic reports/updates easily",
	Argv:    func() interface{} { return new(installT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*installT)
		_ = argv
		return nil
	},
}
