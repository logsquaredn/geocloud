package runtime

import (
	"github.com/logsquaredn/geocloud/datastore"
	"github.com/logsquaredn/geocloud/objectstore"
)

type OSRuntimeOpts struct {
	Datastore   *datastore.Postgres
	Objectstore *objectstore.S3
	WorkDir     string
}
