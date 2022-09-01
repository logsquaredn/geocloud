package main

import (
	"context"
	"os"

	"github.com/logsquaredn/rototiller/pkg/command/ctl"
)

func main() {
	if err := command.NewRoot().ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
