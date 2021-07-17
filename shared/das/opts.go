package das

type DasOpt func (d *Das)

func WithRetries(retries int) DasOpt {
	return func(d *Das) {
		d.retries = retries
	}
}
