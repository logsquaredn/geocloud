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

	"github.com/gin-gonic/gin"
	command "github.com/logsquaredn/rototiller/pkg/command/proxy"

	_ "github.com/logsquaredn/rototiller/pkg/docs/proxy"
	_ "github.com/logsquaredn/rototiller/pkg/service"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	if err := command.NewRoot().ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
