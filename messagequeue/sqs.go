package messagequeue

import (
	"math/rand"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/logsquaredn/geocloud"
	"github.com/rs/zerolog/log"
	"github.com/tedsuo/ifrit"
)

type SQSMessageQueue struct {
	VisibilityTimeout time.Duration `long:"visibility-timeout" default:"15s" description:"Visibilty timeout for SQS messages"`
	QueueNames        []string      `long:"queue-name" description:"SQS queue name"`

	sess  *session.Session
	svc   *sqs.SQS
	rt    geocloud.Runtime
	ds    geocloud.Datastore
	tasks []geocloud.TaskType
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
	q.svc = sqs.New(q.sess)

	wait := make(chan error, 1)
	if len(q.tasks) > 0 && q.ds.IsConfigured() {
		var err error
		tasks, err := q.ds.GetTasks(q.tasks...)
		if err != nil && len(q.QueueNames) == 0 {
			return err
		}

		q.QueueNames = make([]string, len(tasks))
		for i, task := range tasks {
			q.QueueNames[i] = task.QueueID
		}
	}

	if len(q.QueueNames) > 0 {
		queueURLs := []string{}
		for _, name := range q.QueueNames {
			output, err := q.svc.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: &name})
			if err != nil {
				return err
			}
			queueURLs = append(queueURLs, *output.QueueUrl)
		}

		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(queueURLs), func(i, j int) {
			queueURLs[i], queueURLs[j] = queueURLs[j], queueURLs[i]
		})

		if q.VisibilityTimeout < minVisibilityTimeout {
			q.VisibilityTimeout = minVisibilityTimeout
		} else if q.VisibilityTimeout > maxVisibilityTimeout {
			q.VisibilityTimeout = maxVisibilityTimeout
		}

		visticker := time.NewTicker(q.VisibilityTimeout)
		vistimeout := int64(q.VisibilityTimeout.Seconds())
		qticker := time.NewTicker(queueRefreshInterval)

		go func() {
			defer visticker.Stop()
			defer qticker.Stop()
			for i := 0; len(queueURLs) > 0; i = (i + 1) % len(queueURLs) {
				url := queueURLs[i]
				output, err := q.svc.ReceiveMessage(&sqs.ReceiveMessageInput{
					MaxNumberOfMessages: &maxNumberOfMessages,
					QueueUrl:            &url,
				})
				if err != nil {
					wait <- err
				}

				messages := output.Messages
				m := len(messages)
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
							var err error
							tasks, err := q.ds.GetTasks(q.tasks...)
							if err != nil {
								log.Err(err).Msg("unable to update task queue urls")
							}

							q.QueueNames = make([]string, len(tasks))
							for i, task := range tasks {
								q.QueueNames[i] = task.QueueID
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
	return q != nil && q.sess != nil && q.ds.IsConfigured()
}

func (q *SQSMessageQueue) Send(m geocloud.Message) error {
	task, err := q.ds.GetTaskByJobID(m)
	if err != nil {
		return err
	}

	o, err := q.svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &task.QueueID,
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

func (q *SQSMessageQueue) WithSession(sess *session.Session) geocloud.Component {
	q.sess = sess
	return q
}

// TODO refactor polling logic in Run here
// func (q *SQSMessageQueue) poll() error {}
