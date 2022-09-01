package amqp

import "github.com/google/uuid"

const (
	ExchangeName = "rototiller.logsquaredn.io"
)

func NewQueueName(id string) string {
	if id == "" {
		id = uuid.NewString()
	}

	return id + "." + ExchangeName
}
