package api

import "github.com/logsquaredn/geocloud/docs"

// @contact.name   logsquaredn
// @contact.url    https://logsquaredn.io
// @contact.email  logsquaredn@gmail.com

// @license.name  logsquaredn

func init() {
	docs.SwaggerInfo.Title = "Geocloud"
	docs.SwaggerInfo.Description = "Geocloud"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "geocloud.logsquaredn.io"
	docs.SwaggerInfo.BasePath = ""
	docs.SwaggerInfo.Schemes = []string{"https"}
}
