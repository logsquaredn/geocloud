package component

import (
	"os"
	"os/exec"

	"github.com/logsquaredn/geocloud"
	"github.com/tedsuo/ifrit"
)

func NewCmdComponent(cmd *exec.Cmd) geocloud.Component {
	return &cmdComponent{cmd: cmd}
}

type cmdComponent struct {
	cmd *exec.Cmd
}

var _ geocloud.Component = (*cmdComponent)(nil)

func (c *cmdComponent) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	err := c.cmd.Start()
	if err != nil {
		return err
	}

	errs := make(chan error, 1)
	go func() {
		errs <- c.cmd.Wait()
	}()

	close(ready)
	for {
		select {
		case signal := <-signals:
			return c.cmd.Process.Signal(signal)
		case err := <-errs:
			return err
		}
	}
}

func (c *cmdComponent) Execute(_ []string) error {
	return <-ifrit.Invoke(c).Wait()
}

func (c *cmdComponent) Name() string {
	return "cmd"
}

func (c *cmdComponent) IsConfigured() bool {
	return c.cmd != nil
}
