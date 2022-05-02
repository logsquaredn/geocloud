package messagequeue

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/logsquaredn/geocloud"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	"github.com/tedsuo/ifrit"
)

type AMQPMessageQueue struct {
	Enabled    bool          `long:"enabled" description:"Whether or not the AMQP messagequeue is enabled"`
	Address    string        `long:"addr" default:":5432" env:"GEOCLOUD_AMQP_ADDRESS" description:"AMQP address"`
	User       string        `long:"user" default:"geocloud" env:"GEOCLOUD_AMQP_USER" description:"AMQP user"`
	Password   string        `long:"password" env:"GEOCLOUD_AMQP_PASSWORD" description:"AMQP password"`
	Retries    int64         `long:"retries" default:"5" description:"Number of times to retry connecting to AMQP. 0 is infinity"`
	RetryDelay time.Duration `long:"retry-delay" default:"5s" description:"Time to wait between attempts at connecting to AMQP"`
	QueueNames []string      `long:"queue-name" description:"AMQP queue names"`

	conn    *amqp.Connection
	ch      *amqp.Channel
	rt      geocloud.Runtime
	ds      geocloud.Datastore
	tasks   []geocloud.TaskType
	taskmap map[geocloud.TaskType]string
}

var _ geocloud.MessageQueue = (*AMQPMessageQueue)(nil)

func (q *AMQPMessageQueue) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	var (
		err     error
		i       int64 = 1
		connstr       = q.connectionString()
	)
	log.Debug().Msgf("dialing %s", connstr)
	for q.conn, err = amqp.Dial(connstr); err != nil; i++ {
		if i >= q.Retries && q.Retries > 0 {
			return fmt.Errorf("failed to dial amqp after %d attempts: %w", i, err)
		}
		time.Sleep(q.RetryDelay)
	}
	defer q.conn.Close()

	i = 1
	log.Trace().Msg("opening channel")
	for q.ch, err = q.conn.Channel(); err != nil; i++ {
		if i >= q.Retries && q.Retries > 0 {
			return fmt.Errorf("failed to connect to amqp channel after %d attempts: %w", i, err)
		}
		time.Sleep(q.RetryDelay)
	}
	defer q.ch.Close()

	queueNames := q.QueueNames
	for _, name := range queueNames {
		err = q.poll(name)
		if err != nil {
			return err
		}
	}

	qticker := time.NewTicker(queueRefreshInterval)
	defer qticker.Stop()
	close(ready)
	for {
		select {
		case <-qticker.C:
			log.Trace().Msg("refreshing queue names from datastore")
			var err error
			tasks, err := q.ds.GetTasks(q.tasks...)
			if err != nil {
				log.Err(err).Msg("unable to update queue names from datastore")
			}

			for _, task := range tasks {
				if name := task.QueueID; name != "" {
					new := true
					for _, queueName := range queueNames {
						if queueName == name {
							new = false
						}
					}
					if new {
						err = q.poll(name)
						if err != nil {
							return err
						}

						queueNames = append(queueNames, name)
					}
				}
			}
		case <-signals:
			return nil
		}
	}
}

func (q *AMQPMessageQueue) Execute(_ []string) error {
	return <-ifrit.Invoke(q).Wait()
}

func (q *AMQPMessageQueue) Name() string {
	return "amqp"
}

func (q *AMQPMessageQueue) IsEnabled() bool {
	return q.Enabled
}

func (q *AMQPMessageQueue) Send(m geocloud.Message) error {
	task, err := q.ds.GetTaskByJobID(m)
	if err != nil {
		return err
	}

	var queueName = task.QueueID
	if queueName == "" {
		queueName = q.taskmap[task.Type]
		if queueName == "" && len(q.QueueNames) == 1 {
			queueName = q.QueueNames[0]
		}
	}

	log.Debug().Str("id", m.ID()).Msgf("publishing message on %s", queueName)
	return q.ch.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(m.ID()),
		},
	)
}

func (q *AMQPMessageQueue) host() string {
	delimiter := strings.Index(q.Address, ":")
	if delimiter < 0 {
		return q.Address
	}
	return q.Address[:delimiter]
}

func (q *AMQPMessageQueue) port() int64 {
	delimiter := strings.Index(q.Address, ":")
	if delimiter < 0 {
		return 5672
	}
	port, _ := strconv.Atoi(q.Address[delimiter+1:])
	return int64(port)
}

func (q *AMQPMessageQueue) connectionString() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/", q.User, q.Password, q.host(), q.port())
}

func (q *AMQPMessageQueue) WithDatastore(ds geocloud.Datastore) geocloud.MessageQueue {
	q.ds = ds
	return q
}

func (q *AMQPMessageQueue) WithMessageRecipient(rt geocloud.Runtime) geocloud.MessageQueue {
	q.rt = rt
	return q
}

func (q *AMQPMessageQueue) WithTasks(ts ...geocloud.TaskType) geocloud.MessageQueue {
	q.tasks = ts
	return q
}

func (q *AMQPMessageQueue) WithTaskmap(tm map[geocloud.TaskType]string) geocloud.MessageQueue {
	q.taskmap = tm
	return q
}

func (q *AMQPMessageQueue) poll(name string) error {
	if q.rt != nil {
		log.Debug().Msgf("declaring queue %s", name)
		queue, err := q.ch.QueueDeclare(
			name,
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return err
		}

		log.Debug().Msgf("consuming queue %s", queue.Name)
		msgs, err := q.ch.Consume(
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

		go func() {
			for msg := range msgs {
				go func(dTag uint64, body string) {
					k, v := "id", string(body)
					log.Trace().Str(k, v).Msg("sending message to runtime")
					err = q.rt.Send(&message{id: v})
					if err != nil {
						log.Err(err).Str(k, v).Msgf("runtime failed to process message")
					}

					q.ch.Ack(dTag, false)
				}(msg.DeliveryTag, string(msg.Body))
			}
		}()
	}

	return nil
}
