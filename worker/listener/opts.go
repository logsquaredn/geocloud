package listener

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SQSListenerOpt func(l *SQSListener)

func WithCallback(callback SQSListenerCallback) SQSListenerOpt {
	return func(l *SQSListener) {
		l.callback = callback
	}
}

func WithCredentials(creds *credentials.Credentials) SQSListenerOpt {
	return func (l *SQSListener)  {
		l.creds = creds
	}
}

func WithHttpClient(client *http.Client) SQSListenerOpt {
	return func(l *SQSListener) {
		l.client = client
	}
}

func WithQueueNames(names ...string) SQSListenerOpt {
	return func(l *SQSListener) {
		l.names = names
	}
}

func WithQueueUrls(urls ...string) SQSListenerOpt {
	return func(l *SQSListener) {
		l.queues = urls
	}
}

func WithRegion(region string) SQSListenerOpt {
	return func(l *SQSListener) {
		l.region = region
	}
}

func WithService(service *sqs.SQS) SQSListenerOpt {
	return func(l *SQSListener) {
		l.svc = service
	}
}

func WithSession(session *session.Session) SQSListenerOpt {
	return func(l *SQSListener) {
		l.sess = session
	}
}

func WithVisibilityTimeout(timeout time.Duration) SQSListenerOpt {
	return func(l *SQSListener) {
		l.vis = timeout
	}
}
