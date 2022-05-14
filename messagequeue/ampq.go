package messagequeue

import (
	"fmt"
	"time"

	"github.com/logsquaredn/geocloud"
	"github.com/streadway/amqp"
)

type AMQP struct {
	queueName string
	conn      *amqp.Connection
	ch        *amqp.Channel
}

func NewAMQP(opts *AMQPOpts) (*AMQP, error) {
	var (
		q = &AMQP{
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

func (a *AMQP) Send(m geocloud.Message) error {
	return a.ch.Publish("", a.queueName, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(m.GetID()),
	})
}

func (a *AMQP) Poll(f func(geocloud.Message) error) error {
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
		go func(m amqp.Delivery) {
			if err := f(message(string(m.Body))); err == nil {
				m.Ack(false)
			}
		}(m)
	}

	return nil
}
