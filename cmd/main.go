package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/logsquaredn/geocloud"
)

func main() {
	var cmd GeocloudCmd

	cmd.Version = func () {
		fmt.Println(geocloud.Version)
		os.Exit(0)
	}

	parser := flags.NewParser(&cmd, flags.HelpFlag)
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
