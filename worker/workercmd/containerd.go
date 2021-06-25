package workercmd

import (
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	"github.com/logsquaredn/geocloud/shared"
	"github.com/tedsuo/ifrit"
)

type Containerd struct {
	NoRun     bool           `long:"no-run" description:"Whether or not to run its own containerd process. If specified, must target an already-running containerd address"`
	Bin       flags.Filename `long:"bin" default:"containerd" description:"Path to a containerd executable"`
	Config    flags.Filename `long:"config" default:"/etc/geocloud/containerd/config.toml" description:"Path to config file"`
	Loglevel  string         `long:"log-level" default:"debug" choice:"trace" choice:"debug" choice:"info" choice:"warn" choice:"error" choice:"fatal" choice:"panic" description:"Containerd log level"`
	Address   flags.Filename `long:"address" default:"/run/geocloud/containerd/containerd.sock" description:"Address for containerd's GRPC server'"`
	Root      flags.Filename `long:"root" default:"/var/lib/geocloud/containerd" description:"Containerd root directory"`
	State     flags.Filename `long:"state" default:"/run/geocloud/containerd" description:"Containerd state directory"`
	Namespace string         `long:"namespace" default:"geocloud" description:"Containerd namespace"`
}

//go:embed "config.toml"
var toml []byte

func (cmd *WorkerCmd) containerd() (ifrit.Runner, error) {
	var (
		bin      = string(cmd.Containerd.Bin)
		address  = string(cmd.Containerd.Address)
		root     = string(cmd.Containerd.Root)
		state    = string(cmd.Containerd.State)
		config   = string(cmd.Containerd.Config)
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
	if config != "" {
		args = append(args, "--config="+config)
	}
	if loglevel != "" {
		args = append(args, "--log-level="+loglevel)
	}
	if address != "" {
		args = append(args, "--address="+address)
	}
	if root != "" {
		args = append(args, "--root="+root)
	}
	if state != "" {
		args = append(args, "--state="+state)
	}

	containerd := exec.Command(bin, args...)
	containerd.Stdout = os.Stdout
	containerd.Stderr = os.Stderr

	runner, err := shared.NewCmdRunner(shared.WithCmd(containerd))
	if err != nil {
		return nil, err
	}

	return runner, nil
}
