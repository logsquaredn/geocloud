package messagequeue

import (
	"fmt"
	"time"

	"github.com/logsquaredn/geocloud"
	"github.com/streadway/amqp"
)

type amqpMessageQueue struct {
	queueName string
	conn      *amqp.Connection
	ch        *amqp.Channel
}

var _ geocloud.MessageQueue = (*amqpMessageQueue)(nil)

func NewAMQP(opts *AMQPMessageQueueOpts) (*amqpMessageQueue, error) {
	var (
		q = &amqpMessageQueue{
			queueName: opts.QueueName,
		}
		err error
		i   int64 = 1
	)
	for q.conn, err = amqp.Dial(opts.connectionString()); err != nil; i++ {
		if i >= opts.Retries && opts.Retries > 0 {
			return nil, fmt.Errorf("failed to dial amqp after %d attempts: %w", i, err)
		}
		time.Sleep(opts.RetryDelay)
	}

	i = 1
	for q.ch, err = q.conn.Channel(); err != nil; i++ {
		if i >= opts.Retries && opts.Retries > 0 {
			return nil, fmt.Errorf("failed to connect to amqp channel after %d attempts: %w", i, err)
		}
		time.Sleep(opts.RetryDelay)
	}

	return q, nil
}

func (a *amqpMessageQueue) Send(m geocloud.Message) error {
	return a.ch.Publish("", a.queueName, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(m.GetID()),
	})
}

func (a *amqpMessageQueue) Poll(f func(geocloud.Message) error) error {
	queue, _ := a.ch.QueueDeclare(
		a.queueName,
		false,
		false,
		false,
		false,
		nil,
	)

	msgs, err := a.ch.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for m := range msgs {
		if err := f(&message{
			id: string(m.Body),
		}); err == nil {
			if err = m.Ack(false); err != nil {
				return err
			}
		}
	}

	return nil
}
