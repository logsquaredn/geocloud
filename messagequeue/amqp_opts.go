package messagequeue

import (
	"fmt"
	"time"
)

type AMQPMessageQueueOpts struct {
	Host       string
	Port       int64
	User       string
	Password   string
	Retries    int64
	RetryDelay time.Duration
	QueueName  string
}

func (q *AMQPMessageQueueOpts) connectionString() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/", q.User, q.Password, q.Host, q.Port)
}
