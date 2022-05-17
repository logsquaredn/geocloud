package worker

import (
	"github.com/logsquaredn/geocloud/datastore"
	"github.com/logsquaredn/geocloud/objectstore"
)

type Opts struct {
	Datastore   *datastore.Postgres
	Objectstore *objectstore.S3
	WorkDir     string
}
