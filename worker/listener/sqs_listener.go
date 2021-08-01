package listener

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/logsquaredn/geocloud/shared/das"
	"github.com/logsquaredn/geocloud/worker"
	"github.com/rs/zerolog/log"
)

type f = map[string]interface{}

type SQSListenerCallback func (worker.Message) error

type SQSListener struct {
	svc      *sqs.SQS
	das      *das.Das
	tasks    []string
	queues   []string
	vis      time.Duration
	callback SQSListenerCallback
}

var _ worker.Listener = (*SQSListener)(nil)

const runner = "SQSListener"

func New(sess *session.Session, callback SQSListenerCallback, das *das.Das, opts ...SQSListenerOpt) (*SQSListener, error) {
	if callback == nil {
		return nil, fmt.Errorf("listener: nil callback")
	}

	if sess == nil {
		return nil, fmt.Errorf("listener: nil session")
	}

	if das == nil {
		return nil, fmt.Errorf("listener: nil das")
	}
	
	l := &SQSListener{}
	for _, opt := range opts {
		opt(l)
	}

	l.callback = callback
	l.svc = sqs.New(sess)
	l.das = das

	var err error
	l.queues, err = l.getQueueURLs()
	if err != nil {
		return nil, fmt.Errorf("listener: getting queue URLs: %w", err)
	}

	return l, nil
}

var maxNumberOfMessages int64 = 10

func (r *SQSListener) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	q := len(r.queues)
	if q == 0 {
		log.Warn().Fields(f{ "runner": runner }).Msg("no queues specified")
	}

	log.Debug().Fields(f{ "runner": runner }).Msgf("received visibility %ds", int64(r.vis.Seconds()))
	visibility := r.getVisibility()
	log.Info().Fields(f{ "runner": runner }).Msgf("using visibility %ds", int64(visibility.Seconds()))
	vticker := time.NewTicker(visibility)
	timeout := int64(visibility.Seconds())
	d, _ := time.ParseDuration(m5)
	qticker := time.NewTicker(d)

	wait := make(chan error, 1)
	go func() {
		defer vticker.Stop()
		for i := 0; q > 0; i = (i+1)%q {
			url := r.queues[i]
			log.Debug().Fields(f{ "runner": runner, "url": url }).Msg("receiving messages from queue")
			output, err := r.svc.ReceiveMessage(&sqs.ReceiveMessageInput{
				MaxNumberOfMessages: &maxNumberOfMessages,
				QueueUrl: &url,
			})
			if err != nil {
				wait<- err
			}

			messages := output.Messages
			m := len(messages)
			entriesVis := make([]*sqs.ChangeMessageVisibilityBatchRequestEntry, m)
			entriesDel := make([]*sqs.DeleteMessageBatchRequestEntry, m)
			for i, msg := range messages {
				entriesVis[i] = &sqs.ChangeMessageVisibilityBatchRequestEntry{
					Id: msg.MessageId,
					ReceiptHandle: msg.ReceiptHandle,
					VisibilityTimeout: &timeout,
				}

				entriesDel[i] = &sqs.DeleteMessageBatchRequestEntry{
					Id: msg.MessageId,
					ReceiptHandle: msg.ReceiptHandle,
				}
			}

			done := make(chan struct{}, 1)
			go func() {
				log.Debug().Fields(f{ "runner": runner }).Msg("processing messages")
				for _, msg := range messages {
					err = r.callback(&SQSMessage{ msg })
					if err != nil {
						wait<- err
					}
				}

				close(done)
			}()

			for processing := true; processing; {
				select {
				case <-vticker.C:
					if len(entriesVis) > 0 {
						log.Debug().Fields(f{ "runner": runner, "url": url }).Msg("changing messages visibility")
						_, err = r.svc.ChangeMessageVisibilityBatch(&sqs.ChangeMessageVisibilityBatchInput{
							Entries: entriesVis,
							QueueUrl: &url,
						})
						if err != nil {
							wait<- err
						}
					}
				case <-qticker.C:
					log.Info().Fields(f{ "runner": runner }).Msg("updating queue urls")
					r.queues, err = r.getQueueURLs()
					if err != nil {
						wait<- err
					}
					q = len(r.queues)
				case <-done:
					log.Debug().Fields(f{ "runner": runner }).Msg("done processing")
					processing = false
					if len(entriesDel) > 0 {
						log.Debug().Fields(f{ "runner": runner, "url": url }).Msg("deleting messages")
						r.svc.DeleteMessageBatchRequest(&sqs.DeleteMessageBatchInput{
							Entries: entriesDel,
							QueueUrl: &url,
						})
					}
				}
			}
		}
	}()

	log.Debug().Fields(f{ "runner": runner }).Msg("ready")
	close(ready)
	select {
	case err := <-wait:
		log.Err(err).Fields(f{ "runner": runner }).Msg("received error")
		return err
	case signal := <-signals:
		log.Debug().Fields(f{ "runner": runner, "signal": signal.String() }).Msg("received signal")
		return nil
	}
}

const (
	h12 = "12h"
	m5 = "5m"
)

func (l *SQSListener) getVisibility() time.Duration {
	max, _ := time.ParseDuration(h12)
	min, _ := time.ParseDuration(m5)

	if l.vis < min {
		l.vis = min
	} else if l.vis > max {
		l.vis = max
	}

	return l.vis
}

func (l *SQSListener) getQueueURLs() (queues []string, err error) {
	names, err := l.das.GetQueueNamesByTaskTypes(l.tasks...)
	if err != nil {
		return
	}

	for _, name := range names {
		output, err := l.svc.GetQueueUrl(&sqs.GetQueueUrlInput{ QueueName: &name })
		if err != nil {
			return queues, err
		}

		queues = append(queues, *output.QueueUrl)
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(queues), func(i, j int) {
		queues[i], queues[j] = queues[j], queues[i]
	})

	return
}
