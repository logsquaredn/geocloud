package main

import (
	"fmt"
	"os"

	"github.com/logsquaredn/geocloud"
	"github.com/logsquaredn/geocloud/worker/workercmd"
	"github.com/logsquaredn/geocloud/shared"
	"github.com/jessevdk/go-flags"
)

func main() {
	var cmd workercmd.WorkerCmd
	cmd.Version = func() {
		_, err := fmt.Println(geocloud.Version)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	shared.SetupLogging()
	parser := flags.NewParser(&cmd, flags.HelpFlag)
	parser.NamespaceDelimiter = "-"
	args, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			fmt.Println(err)
		} else {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
	}
	cmd.Execute(args)
	os.Exit(0)
}
