package main

import (
	"fmt"
	"os"

	"github.com/mkideal/cli"
)

type addT struct {
	cli.Helper
	File string `cli:"*f,file" usage:"Specify the file to read the logs from"`
	Host string `cli:"*t,host" usage:"Specify the host to send the data to"`
}

var reportCMD = &cli.Command{
	Name:    "report",
	Aliases: []string{"r", "report", "repo"},
	Desc:    "Reports all changes",
	Argv:    func() interface{} { return new(addT) },
	Fn: func(ctx *cli.Context) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Couldn't retrieve homeDir!")
			return nil
		}
		_ = homeDir

		//examle for reading the file line by line
		// err := iptablesparser.ParseFileByLines("/var/log/Tripwire21", func(log *iptablesparser.LogEntry) {
		// 	fmt.Println(*log)
		// })
		// if err != nil {
		// 	panic(err)
		// }
		return nil
	},
}
