package amqp

import (
	"context"
	"encoding/json"

	"github.com/logsquaredn/rototiller"
	"github.com/rabbitmq/amqp091-go"
)

func (a *EventStreamProducer) Emit(ctx context.Context, event *rototiller.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return a.Channel.PublishWithContext(ctx, ExchangeName, event.GetType(), false, false, amqp091.Publishing{
		Body: body,
	})
}
