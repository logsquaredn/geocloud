package rototiller

import "runtime/debug"

var Semver = "0.0.0"

func GetSemver() string {
	semver := Semver

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		var (
			revision string
			modified bool
		)
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				revision = setting.Value
			case "vcs.modified":
				modified = setting.Value == "true"
			}
		}

		if revision != "" {
			i := len(revision)
			if i > 7 {
				i = 7
			}
			semver += "+" + revision[:i]
		}

		if modified {
			semver += "*"
		}
	}

	return semver
}
