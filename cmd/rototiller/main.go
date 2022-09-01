// @contact.name   logsquaredn
// @contact.url    https://rototiller.logsquaredn.io/
// @contact.email  logsquaredn@gmail.com

// @license.name  logsquaredn

package main

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller/pkg/command/d"

	_ "github.com/logsquaredn/rototiller/docs"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/s3blob"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	if err := command.NewRoot().ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
