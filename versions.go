package geocloud

import "fmt"

var (
	Version = "0.1.0"

	Prerelease = ""

	Build = ""
)

func Semver() string {
	version := Version
	if Prerelease != "" {
		version = fmt.Sprintf("%s-%s", version, Prerelease)
	}
	if Build != "" {
		version = fmt.Sprintf("%s+%s", version, Build)
	}
	return version
}
