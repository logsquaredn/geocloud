package runtime

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	"github.com/logsquaredn/geocloud"
	"github.com/logsquaredn/geocloud/component"
	"github.com/tedsuo/ifrit"
)

type ContainerdRuntime struct {
	NoRun     bool           `long:"no-run" description:"Whether or not to run a containerd process. If true, must target an already-running containerd address"`
	Bin       flags.Filename `long:"bin" default:"containerd" description:"Path to a containerd binary"`
	Config    flags.Filename `long:"config" default:"/etc/containerd/config.toml" description:"Path to config file"`
	Loglevel  string         `long:"log-level" default:"info" choice:"trace" choice:"debug" choice:"info" choice:"warn" choice:"error" choice:"fatal" choice:"panic" description:"Containerd log level"`
	Address   flags.Filename `long:"address" default:"/run/containerd/containerd.sock" description:"Address for containerd's gRPC server'"`
	Root      flags.Filename `long:"root" default:"/var/lib/containerd" description:"Containerd root directory"`
	State     flags.Filename `long:"state" default:"/run/containerd" description:"Containerd state directory"`
	Namespace string         `long:"namespace" default:"geocloud" description:"Containerd namespace"`

	ds geocloud.Datastore
	os geocloud.Objectstore

	workdir string
}

var _ geocloud.Runtime = (*ContainerdRuntime)(nil)

//go:embed "config.toml"
var toml []byte

func (c *ContainerdRuntime) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	var (
		bin      = string(c.Bin)
		address  = string(c.Address)
		root     = string(c.Root)
		state    = string(c.State)
		config   = string(c.Config)
		loglevel = c.Loglevel
	)

	if c.workdir == "" {
		c.workdir = os.TempDir()
	}

	if err := os.MkdirAll(filepath.Dir(c.workdir), 0755); err != nil {
		return err
	}

	if _, err := os.Stat(config); err != nil {
		if err := os.MkdirAll(filepath.Dir(config), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(config, toml, 0755); err != nil {
			return err
		}
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

	return component.NewCmdComponent(containerd).Run(signals, ready)
}

func (c *ContainerdRuntime) Execute(_ []string) error {
	return <-ifrit.Invoke(c).Wait()
}

func (c *ContainerdRuntime) Name() string {
	return "containerd"
}

func (c *ContainerdRuntime) IsConfigured() bool {
	return c != nil && c.ds.IsConfigured() && c.os.IsConfigured()
}

func (c *ContainerdRuntime) Send(m geocloud.Message) error {
	return fmt.Errorf("not implemented")
}

func (c *ContainerdRuntime) WithMessageRecipient(_ geocloud.Runtime) geocloud.Runtime {
	// noop
	return c
}

func (c *ContainerdRuntime) WithDatastore(ds geocloud.Datastore) geocloud.Runtime {
	c.ds = ds
	return c
}

func (c *ContainerdRuntime) WithObjectstore(os geocloud.Objectstore) geocloud.Runtime {
	c.os = os
	return c
}
