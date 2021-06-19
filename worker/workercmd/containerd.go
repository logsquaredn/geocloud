package workercmd

import (
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/logsquaredn/geocloud/shared"
)

//go:embed "config.toml"
var toml []byte

func (cmd *WorkerCmd) containerd() (*shared.CmdRunner, error) {
	var (
		bin = string(cmd.Containerd.Bin)
		address = string(cmd.Containerd.Address)
		root = string(cmd.Containerd.Root)
		state = string(cmd.Containerd.State)
		config = string(cmd.Containerd.Config)
		loglevel = cmd.Containerd.Loglevel
	)

	if err := os.MkdirAll(filepath.Dir(config), 0755); err != nil {
		return nil, err
	}

	if err := os.WriteFile(config, toml, 0755); err != nil {
		return nil, err
	}

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

	return &shared.CmdRunner{
		Cmd: containerd,
	}, nil
}
