package main

import (
	"os"

	"github.com/e-travel/cloudwatchlogsbeat/beater"

	"github.com/elastic/beats/libbeat/beat"
)

func main() {
	err := beat.Run("cloudwatchlogsbeat", "", beater.New)
	if err != nil {
		os.Exit(1)
	}
}
