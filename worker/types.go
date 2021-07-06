package worker

import "github.com/tedsuo/ifrit"

type Message interface {
	ID() string
}

type Listener interface {
	ifrit.Runner
}

type Aggregator interface {
	ifrit.Runner

	Aggregate(Message) error
}
