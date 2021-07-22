package janitor

import (
	"fmt"
	"os"

	"github.com/logsquaredn/geocloud/shared/das"
	"github.com/rs/zerolog/log"
	"github.com/tedsuo/ifrit"
)

type f = map[string]interface{}

type Janitor struct {
	das *das.Das
}

func New(das *das.Das, opts... JanitorOpt) (*Janitor, error) {
	if das == nil {
		return nil, fmt.Errorf("janitor: nil das")
	}

	j := &Janitor{}
	for _, opt := range opts {
		opt(j)
	}

	j.das = das

	return j, nil
}

var _ ifrit.Runner = (*Janitor)(nil)

const runner = "Janitor"

func (j *Janitor) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	log.Debug().Fields(f{ "runner":runner }).Msg("ready")
	close(ready)
	<-signals
	return nil
}
