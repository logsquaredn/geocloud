package component

import (
	"os"
	"os/exec"

	"github.com/logsquaredn/geocloud"
	"github.com/tedsuo/ifrit"
)

// NewCmdComponent returns a component that, when ran, executes
// an *exec.Cmd, by making a shell call to some binary
func NewCmdComponent(cmd *exec.Cmd) geocloud.Component {
	return &cmdComponent{cmd: cmd}
}

type cmdComponent struct {
	cmd *exec.Cmd
}

var _ geocloud.Component = (*cmdComponent)(nil)

// Run executes a shell call to an *exec.Cmd and waits for it to return,
// sending it any signals it receives
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

// Execute runs the component and makes it implement the
// flags.Commander interface which lets it be a subcommand
func (c *cmdComponent) Execute(_ []string) error {
	return <-ifrit.Invoke(c).Wait()
}

// Name returns the name of the component to be used to
// identify it when running as a part of a group of components
//
// Do not group a cmdComponent with another cmdComponent
func (c *cmdComponent) Name() string {
	return "cmd"
}

// IsEnabled reports whether or not the component is
// fully configured to be ran
func (c *cmdComponent) IsEnabled() bool {
	return c.cmd != nil
}
