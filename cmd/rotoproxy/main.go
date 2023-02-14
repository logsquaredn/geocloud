// @title        Rototiller
// @version      1.0
// @description  Geospatial data transformation service.

// @contact.name   logsquaredn
// @contact.url    https://rototiller.logsquaredn.io/
// @contact.email  rototiller@logsquaredn.io

// @schemes  https
// @host     rototiller.logsquaredn.io

// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        Authorization

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	command "github.com/logsquaredn/rototiller/command/proxy"

	_ "github.com/logsquaredn/rototiller/api"
	_ "github.com/logsquaredn/rototiller/internal/docs/proxy"
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
