package runtime

import "github.com/logsquaredn/geocloud"

type OSRuntimeOpts struct {
	Datastore   geocloud.Datastore
	Objectstore geocloud.Objectstore
	WorkDir     string
}
