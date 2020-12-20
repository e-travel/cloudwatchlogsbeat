package main

import (
	"os"

	"github.com/e-travel/cloudwatchlogsbeat/beater"
	cmd "github.com/elastic/beats/v7/libbeat/cmd"
	"github.com/elastic/beats/v7/libbeat/cmd/instance"
)

// Name of this beat
var Name = "cloudwatchlogsbeat"

// RootCmd to handle beats cli
var RootCmd = cmd.GenRootCmdWithSettings(beater.New, instance.Settings{Name: Name})

func main() {

	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
