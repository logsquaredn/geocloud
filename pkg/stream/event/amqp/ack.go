package amqp

import "github.com/logsquaredn/rototiller/pkg/api"

func (e *EventStream) Ack(event *api.Event) error {
	return e.Channel.Ack(uint64(event.GetId()), false)
}

func (e *EventStream) Nack(event *api.Event) error {
	return e.Channel.Nack(uint64(event.GetId()), false, true)
}
