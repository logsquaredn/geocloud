package listener

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/logsquaredn/geocloud/worker"
	"github.com/rs/zerolog/log"
)

type f = map[string]interface{}

type SQSListenerCallback func (worker.Message) error

type SQSListener struct {
	svc      *sqs.SQS
	names    []string
	queues   []string
	vis      time.Duration
	callback SQSListenerCallback
}

var _ worker.Listener = (*SQSListener)(nil)

const runner = "SQSListener"

func New(sess *session.Session, callback SQSListenerCallback, opts ...SQSListenerOpt) (*SQSListener, error) {
	if callback == nil {
		return nil, fmt.Errorf("listener: nil callback")
	}

	if sess == nil {
		return nil, fmt.Errorf("listener: nil session")
	}
	
	l := &SQSListener{}
	for _, opt := range opts {
		opt(l)
	}

	l.callback = callback
	l.svc = sqs.New(sess)

	return l, nil
}

var maxNumberOfMessages int64 = 10

func (r *SQSListener) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	log.Debug().Fields(f{ "runner": runner }).Msg("converting queue names to urls")
	for _, name := range r.names {
		output, err := r.svc.GetQueueUrl(&sqs.GetQueueUrlInput{ QueueName: &name })
		if err != nil {
			return err
		}

		r.queues = append(r.queues, *output.QueueUrl)
	}
	
	q := len(r.queues)
	if q == 0 {
		log.Warn().Fields(f{ "runner": runner }).Msg("no queues specified")
	}

	log.Debug().Fields(f{ "runner": runner }).Msg("shuffling queues")
	r.shuffle()

	log.Debug().Fields(f{ "runner": runner }).Msgf("given visibility %ds", int64(r.vis.Seconds()))
	visibility := r.visibility()
	log.Debug().Fields(f{ "runner": runner }).Msgf("using visibility %ds", visibility)

	ticker := r.ticker()

	wait := make(chan error, 1)
	go func() {
		defer ticker.Stop()
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
					VisibilityTimeout: &visibility,
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
				case <-ticker.C:
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

func (l *SQSListener) shuffle() {
	q := len(l.queues)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(q, func(i, j int) {
		l.queues[i], l.queues[j] = l.queues[j], l.queues[i]
	})
}

const (
	max = 43200 // 12h (AWS specified max)
	min = 300 // 5m
)

func (l *SQSListener) visibility() int64 {
	var visibility = int64(l.vis)
	if visibility < min {
		return min
	} else if visibility > max {
		return max
	}

	return visibility
}

func (l *SQSListener) ticker() *time.Ticker {
	return time.NewTicker(l.vis)
}
