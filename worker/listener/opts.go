package listener

import "time"

type SQSListenerOpt func(l *SQSListener)

func WithTasks(tasks ...string) SQSListenerOpt {
	return func(l *SQSListener) {
		l.tasks = append(l.tasks, tasks...)
	}
}

func WithVisibilityTimeout(timeout time.Duration) SQSListenerOpt {
	return func(l *SQSListener) {
		l.vis = timeout
	}
}
