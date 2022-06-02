package main

import (
	"fmt"

	"github.com/logsquaredn/geocloud/internal/conf"
)

var (
	envVarAccessKeyID     = fmt.Sprintf("%sACCESS_KEY_ID", conf.EnvPrefix)
	envVarSecretAccessKey = fmt.Sprintf("%sACCESS_KEY_ID", conf.EnvPrefix)
)

func init() {
	conf.Coil(nil)
}
