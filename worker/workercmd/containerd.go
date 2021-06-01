package workercmd

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/logsquaredn/geocloud/shared"
)

//go:embed "config.toml"
var toml []byte

func (cmd *WorkerCmd) containerd() (*shared.ContainerdRunner, error) {
	var config = string(cmd.Containerd.Config)

	if err := os.MkdirAll(filepath.Dir(config), 0755); err != nil {
		return nil, err
	}

	if err := os.WriteFile(config, toml, 0755); err != nil {
		return nil, err
	}

	return shared.NewContainerdRunner(
		cmd.Containerd.Bin,
		string(cmd.Containerd.Address),
		string(cmd.Containerd.Root),
		string(cmd.Containerd.State),
		config,
		cmd.Containerd.Loglevel,
	)
}
