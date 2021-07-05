package shared

import (
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

type CmdRunnerOpt func (c *CmdRunner)

const runner = "CmdRunner"

func NewCmdRunner(opts ...CmdRunnerOpt) (*CmdRunner, error) {
	c := &CmdRunner{}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

func WithCmd(cmd *exec.Cmd) CmdRunnerOpt {
	return func (c *CmdRunner) {
		c.cmd = cmd
	}
}

func (r *CmdRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	log.Debug().Fields(f{ "runner":runner }).Msg("starting cmd")
	err := r.cmd.Start()
	if err != nil {
		return err
	}

	wait := make(chan error, 1)
	go func() {
		log.Debug().Fields(f{ "runner":runner }).Msg("waiting on cmd to return")
		wait<- r.cmd.Wait()
	}()

	log.Debug().Fields(f{ "runner":runner }).Msg("ready")
	close(ready)
	for {
		select {
		case signal := <-signals:
			log.Debug().Fields(f{ "runner":runner, "signal":signal.String() }).Msg("cmd processing signal")
			return r.cmd.Process.Signal(signal)
		case err := <-wait:
			log.Error().Err(err).Fields(f{ "runner":runner }).Msgf("received error from cmd")
			return err
		}
	}
}
