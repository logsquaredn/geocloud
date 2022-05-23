package main

import (
	"strings"

	"github.com/spf13/viper"
)

func init() {
	viper.SetEnvPrefix("geocloud")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}
