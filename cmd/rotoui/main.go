package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	command "github.com/logsquaredn/rototiller/command/ui"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	if err := command.New().ExecuteContext(ctx); err != nil {
		stop()
		os.Stdout.WriteString(err.Error() + "\n")
		os.Exit(1)
	}

	stop()
	os.Exit(0)
}