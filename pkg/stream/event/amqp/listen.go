package amqp

import (
	"context"
	"encoding/json"

	"github.com/logsquaredn/rototiller"
)

func (a *EventStreamConsumer) Listen(ctx context.Context) (<-chan *rototiller.Event, <-chan error) {
	var (
		eventC = make(chan *rototiller.Event)
		errC   = make(chan error, 1)
	)
	go func() {
		deliveries, err := a.Channel.Consume(a.Queue.Name, "", false, false, false, false, nil)
		if err != nil {
			errC <- err
		}

		for {
			select {
			case delivery := <-deliveries:
				event := &rototiller.Event{}
				if err = json.Unmarshal(delivery.Body, event); err != nil {
					errC <- err
				}
				event.Id = int64(delivery.DeliveryTag)

				eventC <- event
			case <-ctx.Done():
				if err := ctx.Err(); err != nil {
					errC <- err
				}

				close(eventC)
				close(errC)
				break
			}
		}
	}()

	return eventC, errC
}
