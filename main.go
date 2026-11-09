package main

import (
	"nsgo/build"
	"nsgo/flags"
	"nsgo/msg"
	"nsgo/tools"
	"os"
)

func main() {
	msg.Check()
	err := build.Init()
	if err != nil {
		if err.Code() == tools.JustHelp {
			msg.SUCCESSln("help guide provided")
			os.Exit(0)
		}
		msg.ERRORln(err)
		if err.Code() != tools.ErrFileNotFound {
			flags.PrintHelp()
		}
		os.Exit(1)
	}
	if build.Testing {
		msg.INFOln("testing...")
		if msg.IsERROR(build.Tests()) != nil {
			os.Exit(1)
		}
	} else {
		if msg.IsERROR(build.Build()) != nil {
			os.Exit(1)
		}
	}
	msg.SUCCESSln(build.Success)
}
