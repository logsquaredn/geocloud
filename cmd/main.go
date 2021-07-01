package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/logsquaredn/geocloud"
	"github.com/logsquaredn/geocloud/shared"
)

func main() {
	var cmd GeocloudCmd
	version := func() {
		_, err := fmt.Println(geocloud.Version)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	cmd.Version = version
	cmd.API.Version = version
	cmd.Worker.Version = version

	shared.SetupLogging()
	parser := flags.NewParser(&cmd, flags.HelpFlag)
	parser.NamespaceDelimiter = "-"
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			fmt.Println(err)
		} else {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
	}
	os.Exit(0)
}
