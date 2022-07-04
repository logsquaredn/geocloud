// placeholder for swagger
// which is generated at build-time in Dockerfile
// but this much needs to be here because
// github.com/logsquaredn/rototiller/api references it

package docs

import _ "github.com/swaggo/swag" // silence go.mod wanting this to be indirect, as it becomes direct at build time

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

var SwaggerInfo = swaggerInfo{}
