package api

import (
	"github.com/logsquaredn/rototiller/datastore"
	"github.com/logsquaredn/rototiller/messagequeue"
	"github.com/logsquaredn/rototiller/objectstore"
)

type Opts struct {
	Datastore    *datastore.Postgres
	MessageQueue *messagequeue.AMQP
	Objectstore  *objectstore.S3
}
