package worker

import (
	"github.com/logsquaredn/rototiller/datastore"
	"github.com/logsquaredn/rototiller/objectstore"
)

type Opts struct {
	Datastore   *datastore.Postgres
	Objectstore *objectstore.S3
	WorkDir     string
}
