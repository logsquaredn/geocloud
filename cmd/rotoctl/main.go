package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	command "github.com/logsquaredn/rototiller/command/ctl"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	if err := command.New().ExecuteContext(ctx); err != nil {
		stop()
		os.Stdout.WriteString(err.Error() + "\n")
		os.Exit(1)
	}

	stop()
	os.Exit(0)
}
