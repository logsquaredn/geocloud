package oas

type OasOpt func(o *Oas)

func WithPrefix(prefix string) OasOpt {
	return func(o *Oas) {
		o.prefix = prefix
	}
}
