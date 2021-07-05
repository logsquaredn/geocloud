package janitor

import (
	"os"

	"github.com/logsquaredn/geocloud/shared/das"
	"github.com/rs/zerolog/log"
	"github.com/tedsuo/ifrit"
)

type f = map[string]interface{}

type Janitor struct {
	das *das.Das
}

func New(opts... JanitorOpt) (*Janitor, error) {
	return &Janitor{}, nil
}

var _ ifrit.Runner = (*Janitor)(nil)

const runner = "Janitor"

func (j *Janitor) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	log.Debug().Fields(f{ "runner":runner }).Msg("ready")
	close(ready)
	<-signals
	return nil
}
