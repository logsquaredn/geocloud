// placeholder for swagger
// which is generated at build-time in Dockerfile
// but this much needs to be here because
// github.com/logsquaredn/geocloud/api references it

package docs

// silence go.mod wanting this to be
// indirect, as it becomes direct during build
import _ "github.com/swaggo/swag"

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

var SwaggerInfo = swaggerInfo{}
