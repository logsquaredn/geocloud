package janitor

import "github.com/logsquaredn/geocloud/shared/das"

type JanitorOpt func (j* Janitor)

func WithDas(das *das.Das) JanitorOpt {
	return func(j *Janitor) {
		j.das = das
	}
}
