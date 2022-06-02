package conf

import (
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var global = new(globalViper)

type globalViper bool

var _ Coilable = new(globalViper)
var _ Bindable = new(globalViper)

func (g *globalViper) SetEnvPrefix(in string) {
	viper.SetEnvPrefix(in)
}

func (g *globalViper) AutomaticEnv() {
	viper.AutomaticEnv()
}

func (g *globalViper) SetEnvKeyReplacer(r *strings.Replacer) {
	viper.SetEnvKeyReplacer(r)
}

func (g *globalViper) BindPFlag(key string, flag *pflag.Flag) error {
	return viper.BindPFlag(key, flag)
}
