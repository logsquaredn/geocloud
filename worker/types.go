package worker

import (
	"context"

	"github.com/logsquaredn/geocloud"
	"github.com/tedsuo/ifrit"
)

type Listener interface {
	ifrit.Runner
}

type Aggregator interface {
	ifrit.Runner

	Aggregate(context.Context, geocloud.Message) error
}
