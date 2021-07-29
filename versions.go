package geocloud

import (
	"fmt"
	"os"
)

// Version is the version of geocloud. Overridden at build time.
var Version = "0.0.0"

// Commit is the commit hash of geocloud. Overridden at build time.
var Commit = ""

// V prints the version of geocloud embedded within the binary at build time and exits.
func V() {
	v := fmt.Sprintf("geocloud %s", Version)
	if Commit != "" {
		v = fmt.Sprintf("%s %s", v, Commit)
	}
	fmt.Println(v)
	os.Exit(0)
}
