package shared

import (
	"os"
	"os/exec"
	"syscall"
)

type ContainerdRunner struct {
	cmd *exec.Cmd
}

func (r *ContainerdRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	err := r.cmd.Start()
	if err != nil {
		return err
	}

	wait := make(chan error, 1)
	go func() {
		wait <- r.cmd.Wait()
	}()

	close(ready)
	for {
		select {
		case signal := <-signals:
			r.cmd.Process.Signal(signal)
		case err := <-wait:
			return err
		}
	}
}

func NewContainerdRunner(bin, address, root, state, config, loglevel string) (*ContainerdRunner, error) {
	if bin == "" {
		bin = "containerd"
	}
	args := []string{}
	if address != "" {
		args = append(args, "--address="+address)
	}
	if root != "" {
		args = append(args, "--root="+root)
	}
	if state != "" {
		args = append(args, "--state="+state)
	}
	if config != "" {
		args = append(args, "--config="+config)
	}
	if loglevel != "" {
		args = append(args, "--log-level="+loglevel)
	}

	containerd := exec.Command(bin, args...)
	containerd.Stdout = os.Stdout
	containerd.Stderr = os.Stderr
	containerd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGKILL,
	}

	return &ContainerdRunner{
		cmd: containerd,
	}, nil
}
