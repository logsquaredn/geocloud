package geocloud

import (
	"fmt"
	"os"
	"runtime"
)

var (
	// Name is the intended name of the binary
	Name = "geocloud"

	// Package is filled at linking time
	Package = "github.com/logsquaredn/geocloud"

	// Version holds the complete version number. Filled in at linking time
	Version = "0.0.0"

	// Revision is filled with the VCS (e.g. git) revision being used to build
	// the program at linking time
	Revision = ""

	// GoVersion is Go tree's version.
	GoVersion = runtime.Version()
)

// V prints the version of geocloud embedded within the binary at linking time and exits
func V() {
	v := fmt.Sprintf("%s%s", Name, Version)
	if Revision != "" {
		v = fmt.Sprintf("%s+%s", v, Revision)
	}
	fmt.Println(v, GoVersion)
	os.Exit(0)
}
