package runtime

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/remotes"
	"github.com/jessevdk/go-flags"
	"github.com/logsquaredn/geocloud"
	"github.com/logsquaredn/geocloud/component"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/rs/zerolog/log"
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

	ctx 	 context.Context
	workdir  string
	client   *containerd.Client
	resolver *remotes.Resolver
}

var _ geocloud.Runtime = (*ContainerdRuntime)(nil)

//go:embed "config.toml"
var toml []byte

func (c *ContainerdRuntime) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	var (
		bin       = string(c.Bin)
		address   = string(c.Address)
		root      = string(c.Root)
		state     = string(c.State)
		config    = string(c.Config)
		loglevel  = c.Loglevel
		namespace = c.Namespace
	)

	c.ctx = namespaces.WithNamespace(context.Background(), namespace)

	if c.workdir == "" {
		c.workdir = os.TempDir()
	}

	if err := os.MkdirAll(c.jobsdir(), 0755); err != nil {
		return err
	}

	if _, err := os.Stat(config); os.IsNotExist(err) {
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

	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return component.NewNamedGroup(
		c.Name(),
		component.NewCmdComponent(cmd),
		component.NewComponentFunc(
			func(sgnls <-chan os.Signal, rdy chan<- struct{}) error {
				var err error
				c.client, err = containerd.New(
					address,
					containerd.WithDefaultNamespace(namespace),
				)
				if err != nil {
					return err
				}

				close(rdy)
				<-sgnls
				return nil
			},
		),
	).Run(signals, ready)
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
	f := map[string]string{ "id": m.ID() }
	log.Info().Fields(f).Msgf("processing message")

	log.Trace().Fields(f).Msgf("getting job from datastore")
	j, err := c.ds.GetJob(m)
	if err != nil {
		return err
	}

	defer func() {
		j.EndTime = time.Now()
		j.Err = err
		if j.Err != nil && j.Err.Error() != "" {
			j.Status = geocloud.Error
		} else {
			j.Status = geocloud.Complete
		}
		log.Info().Fields(f).Msgf("job finished with status %s", j.Status.Status())
		c.ds.UpdateJob(j)
	}()

	j.Status = geocloud.InProgress
	log.Trace().Fields(f).Msgf("settings job to %s", j.Status.Status())
	j, err = c.ds.UpdateJob(j)
	if err != nil {
		return err
	}

	log.Trace().Fields(f).Msgf("getting task for job from datastore")
	t, err := c.ds.GetTaskByJobID(m)
	if err != nil {
		return err
	}

	var (
		image containerd.Image
		imgrdy = make(chan error, 1)
		volrdy = make(chan error, 1)
	)
	go func() {
		log.Debug().Fields(f).Msgf("pulling image %s", t.Ref)
		if image, err = c.pull(t.Ref); err != nil {
			imgrdy<- err
		} else {
			close(imgrdy)
		}
	}()

	log.Trace().Fields(f).Msg("creating input volume")
	invol, err := c.involume(m)
	if err != nil {
		return err
	}
	defer os.RemoveAll(c.jobdir(m))

	log.Trace().Fields(f).Msgf("getting input")
	input, err := c.os.GetInput(m)
	if err != nil {
		return err
	}

	go func() {
		log.Debug().Fields(f).Msgf("downloading input")
		if err = input.Download(invol.path); err != nil {
			volrdy<- err
		} else {
			close(volrdy)
		}
	}()

	log.Trace().Fields(f).Msg("creating output volume")
	outvol, err := c.outvolume(m)
	if err != nil {
		return err
	}

	if err = <-imgrdy; err != nil {
		return err
	}
	log.Debug().Fields(f).Msgf("finished pulling image %s", t.Ref)
	if err = <-volrdy; err != nil {
		return err
	}
	log.Debug().Fields(f).Msgf("finished downloading input")

	var filename string
	invol.Walk(func(_ string, f geocloud.File, e error) error {
		if e != nil {
			return e
		}
		filename = f.Name()
		return fmt.Errorf("found") // we only expect 1 input, so use the first one we find
	})

	if filename == "" {
		return fmt.Errorf("no input found")
	}

	args := append([]string { filepath.Join("/job/input", filename), "/job/output" }, j.Args...)
	mounts := []specs.Mount{
		c.mount(invol.path, filepath.Dir(args[0]), "ro"),
		c.mount(outvol.path, args[1], "rw"),
	}

	var v, a string
	for _, m := range mounts {
		v += fmt.Sprintf(" -v %s:%s", m.Source, m.Destination)
	}
	for _, r := range args {
		a += fmt.Sprintf(" %s", r)
	}

	log.Info().Fields(f).Msgf("running%s %s%s", v, t.Ref, a)
	container, err := c.run(m, image, args, mounts)
	if err != nil {
		return err
	}
	defer container.Delete(c.ctx, containerd.WithSnapshotCleanup)

	log.Trace().Fields(f).Msg("creating task")
	jobErr := []byte{}
	stderr := bytes.NewBuffer(jobErr)
	task, err := container.NewTask(c.ctx, cio.NewCreator(cio.WithStreams(os.Stdin, os.Stdout, stderr)))
 	if err != nil {
 		return err
 	}
 	defer task.Delete(c.ctx)

	log.Trace().Fields(f).Msg("waiting on task to set up")
	exitStatusC, err := task.Wait(c.ctx)
 	if err != nil {
 		return err
 	}

	log.Trace().Fields(f).Msg("starting task")
	if err := task.Start(c.ctx); err != nil {
		return err
	}

	log.Trace().Fields(f).Msg("waiting on task to complete")
	exitStatus := <-exitStatusC
 	if exitCode, _, err := exitStatus.Result(); err != nil {
 		return err
 	} else if exitCode != 0 {
		err = fmt.Errorf("job exited with code %d", exitCode)
		return err
	}

	log.Debug().Fields(f).Msg("uploading output")
	if err = c.os.PutOutput(m, outvol); err != nil {
		return err
	}

	if len(jobErr) > 0 {
		err = fmt.Errorf("%s", jobErr)
	}

	return err
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

func (c *ContainerdRuntime) WithResolver(r *remotes.Resolver) *ContainerdRuntime {
	c.resolver = r
	return c
}

func (c *ContainerdRuntime) WithWorkdir(w string) geocloud.Runtime {
	c.workdir = w
	return c
}

func (c *ContainerdRuntime) pull(ref string) (containerd.Image, error) {
	return c.client.Pull(
		c.ctx, ref,
		containerd.WithPullUnpack,
		containerd.WithResolver(*c.resolver),
	)
}

func volume(path string) (*dirVolume, error) {
	return &dirVolume{ path: path }, os.MkdirAll(path, 0755)
}


func (c *ContainerdRuntime) involume(m geocloud.Message) (*dirVolume, error) {
	return volume(filepath.Join(c.jobdir(m), "input"))
}

func (c *ContainerdRuntime) outvolume(m geocloud.Message) (*dirVolume, error) {
	return volume(filepath.Join(c.jobdir(m), "output"))
}

func (c *ContainerdRuntime) jobdir(m geocloud.Message) string {
	return filepath.Join(c.jobsdir(), m.ID())
}

func (c *ContainerdRuntime) jobsdir() string {
	return filepath.Join(c.workdir, "jobs")
}

func (c *ContainerdRuntime) mount(src, dst string, opts ...string) specs.Mount {
	return specs.Mount{
		Source: src,
		Destination: dst,
		Type: "bind",
		Options: append([]string{ "bind" }, opts...),
	}
}

func (c *ContainerdRuntime) run(m geocloud.Message, image containerd.Image, args []string, mounts []specs.Mount) (containerd.Container, error) {
	return c.client.NewContainer(
		c.ctx, m.ID(),
		containerd.WithNewSnapshot(m.ID(), image),
		containerd.WithNewSpec(
			oci.WithImageConfigArgs(image, args), 
			oci.WithMounts(mounts),
		),
	)
}
