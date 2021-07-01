package worker

import "github.com/tedsuo/ifrit"

type Message interface {
	// wraps sqs.Message so that other message queues can be used in the future
	// by implementing this interface

	ID() string
}

type Listener interface {
	ifrit.Runner
	// listens for SQS messages and sends them onto Aggregator
}

type Aggregator interface {
	ifrit.Runner
	// receives SQS messages and aggregates info from Postgres, S3 about message
	// after compiling everything, sends the request onto be processed

	Aggregate(Message) error
}
