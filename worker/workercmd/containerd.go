package workercmd

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/logsquaredn/geocloud/runners"
	"github.com/tedsuo/ifrit"
)

//go:embed "config.toml"
var toml []byte

func (cmd *WorkerCmd) containerd() (ifrit.Runner, error) {	
	i := runners.ContainerdRunnerInput{
		Bin:      cmd.Containerd.Bin,
		Address:  string(cmd.Containerd.Address),
		Config:   string(cmd.Containerd.Config),
		LogLevel: cmd.Containerd.LogLevel,
		Root:     string(cmd.Containerd.Root),
		State:    string(cmd.Containerd.State),
	}

	if err := os.MkdirAll(filepath.Dir(i.Config), 0755); err != nil {
		return nil, err
	}

	if err := os.WriteFile(i.Config, toml, 0755); err != nil {
		return nil, err
	}

	return runners.NewContainerdRunner(i)
}
