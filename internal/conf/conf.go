package conf

import (
	"time"

	"github.com/spf13/pflag"
)

type Bindable interface {
	BindPFlag(string, *pflag.Flag) error
}

type Conf struct {
	Arg         string
	Default     interface{}
	Description string
}

func BindToFlags(flags *pflag.FlagSet, b Bindable, cs ...*Conf) error {
	if b == nil {
		b = global
	}

	for _, c := range cs {
		switch t := c.Default.(type) {
		case string:
			flags.String(c.Arg, t, c.Description)
		case int64:
			flags.Int64(c.Arg, t, c.Description)
		case bool:
			flags.Bool(c.Arg, t, c.Description)
		case time.Duration:
			flags.Duration(c.Arg, t, c.Description)
		}
		if err := b.BindPFlag(c.Arg, flags.Lookup(c.Arg)); err != nil {
			return err
		}
	}
	return nil
}
