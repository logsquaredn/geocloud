package das

import "time"

type DasOpt func (d *Das)

func WithRetries(retries int) DasOpt {
	return func(d *Das) {
		d.retries = retries
	}
}

func WithRetryDelay(delay time.Duration) DasOpt {
	return func(d *Das) {
		d.delay = delay
	}
}
