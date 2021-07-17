package shared

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"
	"github.com/tedsuo/ifrit"
)

type f = map[string]interface{}

type CmdRunner struct {
	cmd *exec.Cmd
}

var _ ifrit.Runner = (*CmdRunner)(nil)

const runner = "CmdRunner"

func NewCmdRunner(cmd *exec.Cmd) (*CmdRunner, error) {
	return &CmdRunner{
		cmd: cmd,
	}, nil
}

func (r *CmdRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	log.Debug().Fields(f{ "runner": runner }).Msg("starting cmd")
	err := r.cmd.Start()
	if err != nil {
		return fmt.Errorf("shared: failed to start cmd: %w", err)
	}

	wait := make(chan error, 1)
	go func() {
		log.Debug().Fields(f{ "runner": runner }).Msg("waiting on cmd to return")
		wait<- r.cmd.Wait()
	}()

	log.Debug().Fields(f{ "runner": runner }).Msg("ready")
	close(ready)
	for {
		select {
		case signal := <-signals:
			log.Debug().Fields(f{ "runner": runner, "signal": signal.String() }).Msg("cmd processing signal")
			return r.cmd.Process.Signal(signal)
		case err := <-wait:
			log.Error().Err(err).Fields(f{ "runner": runner }).Msg("received error from cmd")
			return fmt.Errorf("shared: received error from cmd: %w", err)
		}
	}
}
