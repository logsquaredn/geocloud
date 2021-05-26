package runners

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/tedsuo/ifrit"
)

type ContainerdRunnerInput struct {
	Address  string
	Bin      string
	Config   string
	LogLevel string
	Root     string
	State    string
}

func NewContainerdRunner(i ContainerdRunnerInput) (ifrit.Runner, error) {
	if err := i.Required(); err != nil {
		return nil, err
	}

	args := []string {
		"--address="+i.Address,
		"--root="+i.Root,
		"--state="+i.State,
	}
	if i.Config != "" {
		args = append(args, "--config="+i.Config)
	}
	if i.LogLevel != "" {
		args = append(args, "--log-level="+i.LogLevel)
	}

	containerd := exec.Command(i.Bin, args...)
	containerd.Stdout = os.Stdout
	containerd.Stderr = os.Stderr
	containerd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGKILL,
	}

	return ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
		err := containerd.Start()
		if err != nil {
			return err
		}

		wait := make(chan error, 1)
		go func() {
			wait <-containerd.Wait()
		}()

		close(ready)

		for {
			select {
			case signal := <-signals:
				containerd.Process.Signal(signal)
			case err := <-wait:
				return err
			}
		}
	}), nil
}

func (i *ContainerdRunnerInput) Required() error {
	args := []string {
		i.Address,
		i.Bin,
		i.Root,
		i.State,
	}
	for _, arg := range args {
		if arg == "" {
			return fmt.Errorf("required arg not specified")
		}
	}
	return nil
}
