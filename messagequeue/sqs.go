package messagequeue

import (
	"math/rand"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/logsquaredn/geocloud"
	"github.com/rs/zerolog/log"
	"github.com/tedsuo/ifrit"
)

type SQSMessageQueue struct {
	VisibilityTimeout time.Duration `long:"visibility-timeout" default:"15s" description:"Visibilty timeout for SQS messages"`
	QueueNames        []string      `long:"queue-name" description:"SQS queue name"`

	cfg   *aws.Config
	svc   *sqs.SQS
	rt    geocloud.Runtime
	ds    geocloud.Datastore
	tasks []geocloud.TaskType
	taskmap map[geocloud.TaskType]string
}

var _ geocloud.AWSComponent = (*SQSMessageQueue)(nil)
var _ geocloud.MessageQueue = (*SQSMessageQueue)(nil)

const (
	t12h = "12h"
	tm5  = "5m"
	t5s  = "5s"
)

var (
	maxNumberOfMessages  int64 = 10
	maxVisibilityTimeout time.Duration
	minVisibilityTimeout time.Duration
	queueRefreshInterval time.Duration
)

func init() {
	maxVisibilityTimeout, _ = time.ParseDuration(t12h)
	minVisibilityTimeout, _ = time.ParseDuration(t5s)
	queueRefreshInterval, _ = time.ParseDuration(tm5)
}

func (q *SQSMessageQueue) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	sess, err := session.NewSession(q.cfg)
	if err != nil {
		return err
	}
	q.svc = sqs.New(sess)

	var queueNames = q.QueueNames
	wait := make(chan error, 1)
	if len(q.tasks) > 0 && q.ds.IsConfigured() {
		log.Trace().Msg("getting tasks from datastore")
		var err error
		tasks, err := q.ds.GetTasks(q.tasks...)
		if err != nil && len(queueNames) == 0 {
			return err
		}

		for _, task := range tasks {
			if task.QueueID != "" {
				queueNames = append(queueNames, task.QueueID)
			}
		}
	}

	if len(queueNames) > 0 {
		log.Trace().Msg("getting queue URLs from queue names")
		queueURLs := []string{}
		for _, name := range queueNames {
			output, err := q.svc.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: &name})
			if err != nil {
				return err
			}
			queueURLs = append(queueURLs, *output.QueueUrl)
		}

		log.Trace().Msg("shuffling queue URLs")
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(queueURLs), func(i, j int) {
			queueURLs[i], queueURLs[j] = queueURLs[j], queueURLs[i]
		})

		if q.VisibilityTimeout < minVisibilityTimeout {
			q.VisibilityTimeout = minVisibilityTimeout
		} else if q.VisibilityTimeout > maxVisibilityTimeout {
			q.VisibilityTimeout = maxVisibilityTimeout
		}
		log.Trace().Msgf("using visibility timeout %d", q.VisibilityTimeout)

		visticker := time.NewTicker(q.VisibilityTimeout)
		vistimeout := int64(q.VisibilityTimeout.Seconds())
		qticker := time.NewTicker(queueRefreshInterval)

		go func() {
			defer visticker.Stop()
			defer qticker.Stop()
			for i := 0; len(queueURLs) > 0; i = (i + 1) % len(queueURLs) {
				url := queueURLs[i]
				log.Trace().Msgf("polling %s", url)
				output, err := q.svc.ReceiveMessage(&sqs.ReceiveMessageInput{
					MaxNumberOfMessages: &maxNumberOfMessages,
					QueueUrl:            &url,
				})
				if err != nil {
					wait <- err
				}

				messages := output.Messages
				m := len(messages)
				log.Trace().Msgf("got %d messages", m)
				entriesVis := make([]*sqs.ChangeMessageVisibilityBatchRequestEntry, m)
				for i, msg := range messages {
					entriesVis[i] = &sqs.ChangeMessageVisibilityBatchRequestEntry{
						Id:                msg.MessageId,
						ReceiptHandle:     msg.ReceiptHandle,
						VisibilityTimeout: &vistimeout,
					}
				}

				entriesDel := []*sqs.DeleteMessageBatchRequestEntry{}
				done := make(chan struct{}, 1)
				go func() {
					for _, msg := range messages {
						k, v := "id", *msg.Body
						log.Trace().Str(k, v).Msg("sending message to runtime")
						err = q.rt.Send(&message{id: v})
						if err != nil {
							log.Err(err).Str(k, v).Msgf("runtime failed to process message")
						} else {
							entriesDel = append(entriesDel, &sqs.DeleteMessageBatchRequestEntry{
								Id:            msg.MessageId,
								ReceiptHandle: msg.ReceiptHandle,
							})
						}
					}

					close(done)
				}()

				for processing := true; processing; {
					select {
					case <-visticker.C:
						if len(entriesVis) > 0 {
							log.Trace().Msg("changing message visibility")
							_, err = q.svc.ChangeMessageVisibilityBatch(&sqs.ChangeMessageVisibilityBatchInput{
								Entries:  entriesVis,
								QueueUrl: &url,
							})
							if err != nil {
								wait <- err
							}
						}
					case <-qticker.C:
						if q.ds.IsConfigured() && len(q.tasks) > 0 {
							log.Trace().Msg("refreshing queue names from datastore")
							var err error
							tasks, err := q.ds.GetTasks(q.tasks...)
							if err != nil {
								log.Err(err).Msg("unable to update task queue urls")
							}

							queueNames = q.QueueNames
							for _, task := range tasks {
								if task.QueueID != "" {
									queueNames = append(queueNames, task.QueueID)
								}
							}

							newQueueURLs := []string{}
							for _, name := range q.QueueNames {
								output, err := q.svc.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: &name})
								if err != nil {
									log.Err(err).Msgf("unable to get %s queue url", name)
								} else {
									newQueueURLs = append(queueURLs, *output.QueueUrl)
								}
							}

							rand.Seed(time.Now().UnixNano())
							rand.Shuffle(len(newQueueURLs), func(i, j int) {
								newQueueURLs[i], newQueueURLs[j] = newQueueURLs[j], newQueueURLs[i]
							})

							queueURLs = newQueueURLs
						}
					case <-done:
						processing = false
						if len(entriesDel) > 0 {
							log.Trace().Msg("deleting messages")
							req, _ := q.svc.DeleteMessageBatchRequest(&sqs.DeleteMessageBatchInput{
								Entries:  entriesDel,
								QueueUrl: &url,
							})
							err = req.Send()
							if err != nil {
								wait <- err
							}
						}
					}
				}
			}
		}()
	}

	close(ready)
	select {
	case err := <-wait:
		return err
	case <-signals:
		return nil
	}
}

func (q *SQSMessageQueue) Execute(_ []string) error {
	return <-ifrit.Invoke(q).Wait()
}

func (q *SQSMessageQueue) Name() string {
	return "sqs"
}

func (q *SQSMessageQueue) IsConfigured() bool {
	return q != nil && q.cfg != nil && q.ds.IsConfigured()
}

func (q *SQSMessageQueue) Send(m geocloud.Message) error {
	task, err := q.ds.GetTaskByJobID(m)
	if err != nil {
		return err
	}

	var queueName = task.QueueID
	if queueName == "" {
		queueName = q.taskmap[task.Type]
	}

	o, err := q.svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})
	if err != nil {
		return err
	}

	id := m.ID()
	_, err = q.svc.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    o.QueueUrl,
		MessageBody: &id,
	})
	return err
}

func (q *SQSMessageQueue) WithDatastore(ds geocloud.Datastore) geocloud.MessageQueue {
	q.ds = ds
	return q
}

func (q *SQSMessageQueue) WithMessageRecipient(rt geocloud.Runtime) geocloud.MessageQueue {
	q.rt = rt
	return q
}

func (q *SQSMessageQueue) WithTasks(ts ...geocloud.TaskType) geocloud.MessageQueue {
	q.tasks = ts
	return q
}

func (q *SQSMessageQueue) WithTaskmap(tm map[geocloud.TaskType]string) geocloud.MessageQueue {
	q.taskmap = tm
	return q
}

func (q *SQSMessageQueue) WithConfig(cfg *aws.Config) geocloud.AWSComponent {
	q.cfg = cfg
	return q
}

// TODO refactor polling logic in Run here
// func (q *SQSMessageQueue) poll() error {}
