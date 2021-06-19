package shared

import (
	"os"
	"os/exec"
)

type CmdRunner struct {
	Cmd *exec.Cmd
}

func (r *CmdRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	err := r.Cmd.Start()
	if err != nil {
		return err
	}

	wait := make(chan error, 1)
	go func() {
		wait <- r.Cmd.Wait()
	}()

	close(ready)
	for {
		select {
		case signal := <-signals:
			r.Cmd.Process.Signal(signal)
		case err := <-wait:
			return err
		}
	}
}
