package api

import (
	"github.com/logsquaredn/geocloud/datastore"
	"github.com/logsquaredn/geocloud/messagequeue"
	"github.com/logsquaredn/geocloud/objectstore"
)

type ServerOpts struct {
	Datastore    *datastore.Postgres
	MessageQueue *messagequeue.AMQP
	Objectstore  *objectstore.S3
}
