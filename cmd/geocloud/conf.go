package main

import (
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type conf struct {
	arg  string
	def  interface{}
	desc string
}

func bindConfToFlags(flags *pflag.FlagSet, cs ...*conf) {
	for _, c := range cs {
		switch t := c.def.(type) {
		case string:
			flags.String(c.arg, t, c.desc)
		case int64:
			flags.Int64(c.arg, t, c.desc)
		case bool:
			flags.Bool(c.arg, t, c.desc)
		case time.Duration:
			flags.Duration(c.arg, t, c.desc)
		}
		_ = viper.BindPFlag(c.arg, flags.Lookup(c.arg))
	}
}
