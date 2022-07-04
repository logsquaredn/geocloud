package api

import "github.com/logsquaredn/rototiller/docs"

// @contact.name   logsquaredn
// @contact.url    https://logsquaredn.io
// @contact.email  logsquaredn@gmail.com

// @license.name  logsquaredn

func init() {
	docs.SwaggerInfo.Title = "Rototiller"
	docs.SwaggerInfo.Description = "Rototiller"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "rototiller.logsquaredn.io"
	docs.SwaggerInfo.BasePath = ""
	docs.SwaggerInfo.Schemes = []string{"https"}
}
