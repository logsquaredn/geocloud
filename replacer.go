package geocloud

import "strings"

var QueryParamToEnvVarReplacer = strings.NewReplacer("-", "_")
