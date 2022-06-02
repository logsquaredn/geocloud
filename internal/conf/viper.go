package conf

import (
	"strings"
)

const (
	ViperEnvPrefix = "geocloud"
	EnvPrefix      = "GEOCLOUD_"
)

type Coilable interface {
	SetEnvPrefix(string)
	AutomaticEnv()
	SetEnvKeyReplacer(*strings.Replacer)
}

func Coil(c Coilable) {
	if c == nil {
		c = global
	}

	c.SetEnvPrefix(ViperEnvPrefix)
	c.AutomaticEnv()
	c.SetEnvKeyReplacer(HyphenToUnderscoreReplacer)
}
