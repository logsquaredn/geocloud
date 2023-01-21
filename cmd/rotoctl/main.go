package main

import (
	"context"
	"os"

	command "github.com/logsquaredn/rototiller/command/ctl"
)

func main() {
	if err := command.NewRoot().ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
