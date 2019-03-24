package main

import (
	"os"
	"strings"

	"github.com/budden/semdict/pkg/app"
)

func main() {
	parseCommandLineFlags()
	app.Run(os.Args)
}

func parseCommandLineFlags() {
	commandName := os.Args[0]
	cfn := app.DefaultConfigFileName
	tbd := "" //templateBaseDir
	if strings.Contains(commandName, "bin/") {
		cfn = "/etc/semdict/" + cfn
		tbd = "/usr/share/semdict/"
	}
	app.ConfigFileName = &cfn
	app.TemplateBaseDir = &tbd
	return
}
