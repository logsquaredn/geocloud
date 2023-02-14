package amqp

import "github.com/logsquaredn/rototiller/pb"

func (e *EventStream) Ack(event *pb.Event) error {
	return e.Channel.Ack(uint64(event.GetId()), false)
}

func (e *EventStream) Nack(event *pb.Event) error {
	return e.Channel.Nack(uint64(event.GetId()), false, true)
}
