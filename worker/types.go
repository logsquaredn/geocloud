package worker

import (
	"github.com/logsquaredn/geocloud"
	"github.com/tedsuo/ifrit"
)

type Listener interface {
	ifrit.Runner
}

type Aggregator interface {
	ifrit.Runner

	Aggregate(geocloud.Message) error
}
