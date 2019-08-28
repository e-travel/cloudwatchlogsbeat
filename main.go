package main

import (
	"os"

	"github.com/e-travel/cloudwatchlogsbeat/beater"

	"github.com/elastic/beats/libbeat/cmd"
)

var RootCmd = cmd.GenRootCmd("cloudwatchlogsbeat", "", beater.New)

func main() {

	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
