package api

import "github.com/logsquaredn/geocloud"

type GinOpts struct {
	Datastore    geocloud.Datastore
	Objectstore  geocloud.Objectstore
	MessageQueue geocloud.MessageRecipient
}
