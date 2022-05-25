package main

import (
	"github.com/logsquaredn/geocloud"
	"github.com/spf13/viper"
)

func init() {
	viper.SetEnvPrefix("geocloud")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(geocloud.QueryParamToEnvVarReplacer)
}
