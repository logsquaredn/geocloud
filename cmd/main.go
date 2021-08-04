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
	cmd.Version = geocloud.V
	cmd.API.Version = geocloud.V
	cmd.Worker.Version = geocloud.V
	cmd.Migrate.Version = geocloud.V

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
