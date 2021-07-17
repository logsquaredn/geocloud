package listener

import "time"

type SQSListenerOpt func(l *SQSListener)

func WithCallback(callback SQSListenerCallback) SQSListenerOpt {
	return func(l *SQSListener) {
		l.callback = callback
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

func WithVisibilityTimeout(timeout time.Duration) SQSListenerOpt {
	return func(l *SQSListener) {
		l.vis = timeout
	}
}
