package amqp

import (
	"context"
	"net/url"
	"os"
	"strings"

	"github.com/logsquaredn/rototiller/pkg/api"
	amqp "github.com/rabbitmq/amqp091-go"
)

type EventStream struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

type EventStreamProducer struct {
	*EventStream
}

type EventStreamConsumer struct {
	*EventStream
	Queue *amqp.Queue
}

func New(ctx context.Context, addr string) (*EventStream, error) {
	if addr == "" {
		addr = os.Getenv("AMQP_ADDR")
	}

	addr = "amqp://" + strings.TrimPrefix(addr, "amqp://")

	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	if u.User.String() == "" {
		u.User = url.UserPassword(os.Getenv("AMQP_USERNAME"), os.Getenv("AMQP_PASSWORD"))
	}

	connection, err := amqp.Dial(u.String())
	if err != nil {
		return nil, err
	}

	channel, err := connection.Channel()
	if err != nil {
		return nil, err
	}

	if err := channel.ExchangeDeclare(ExchangeName, amqp.ExchangeTopic, true, false, false, false, nil); err != nil {
		return nil, err
	}

	return &EventStream{connection, channel}, nil
}

func (e *EventStream) Close() error {
	if err := e.Channel.Close(); err != nil {
		defer e.Connection.Close()
		return err
	}

	return e.Connection.Close()
}

func (e *EventStream) NewProducer(ctx context.Context) (*EventStreamProducer, error) {
	return &EventStreamProducer{e}, nil
}

func (e *EventStream) NewConsumer(ctx context.Context, id string, events ...api.EventType) (*EventStreamConsumer, error) {
	queue, err := e.Channel.QueueDeclare(NewQueueName(id), true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	for _, event := range events {
		if err = e.Channel.QueueBind(queue.Name, event.String(), ExchangeName, false, nil); err != nil {
			return nil, err
		}
	}

	return &EventStreamConsumer{e, &queue}, nil
}

func NewProducer(ctx context.Context, addr string) (*EventStreamProducer, error) {
	e, err := New(ctx, addr)
	if err != nil {
		return nil, err
	}

	return e.NewProducer(ctx)
}

func NewConsumer(ctx context.Context, addr, id string) (*EventStreamConsumer, error) {
	e, err := New(ctx, addr)
	if err != nil {
		return nil, err
	}

	return e.NewConsumer(ctx, id)
}
