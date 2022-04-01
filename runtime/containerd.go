package runtime

import (
	"bytes"
	"context"

	// embed must be imported to use go:embed
	_ "embed"
	"errors"
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
	Retries   int64          `long:"retries" default:"5" description:"Number of times to retry connecting to Containerd. 0 is infinity"`
	Timeout   time.Duration  `long:"timeout" description:"Time to wait between attempts at connecting to Containerd. Containerd defaults to 10s"`

	ds geocloud.Datastore
	os geocloud.Objectstore

	ctx      context.Context
	workdir  string
	client   *containerd.Client
	resolver *remotes.Resolver
}

var _ geocloud.Runtime = (*ContainerdRuntime)(nil)

//go:embed "config.toml"
var toml []byte

// tasks.tar is generated at build time from tasks/ directory
// run `make save-tasks` to suppress this warning locally
//go:embed "tasks.tar"
var tar []byte

func (c *ContainerdRuntime) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	var (
		bin        = string(c.Bin)
		address    = string(c.Address)
		root       = string(c.Root)
		state      = string(c.State)
		config     = string(c.Config)
		loglevel   = c.Loglevel
		namespace  = c.Namespace
		components = []geocloud.Component{}
	)

	c.ctx = namespaces.WithNamespace(context.Background(), namespace)

	if c.workdir == "" {
		c.workdir = os.TempDir()
	}

	if err := os.MkdirAll(c.jobsdir(), 0755); err != nil {
		return err
	}

	if !c.NoRun {
		if _, err := os.Stat(config); errors.Is(err, os.ErrNotExist) {
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

		components = append(components, component.NewCmdComponent(cmd))
	}

	components = append(
		components,
		component.NewComponentFunc(
			func(sgnls <-chan os.Signal, rdy chan<- struct{}) error {
				var (
					err error
					i   int64 = 1
				)
				for c.client, err = containerd.New(
					address,
					containerd.WithDefaultNamespace(namespace),
					containerd.WithTimeout(c.Timeout),
				); err != nil; i++ {
					if i >= c.Retries && c.Retries > 0 {
						return fmt.Errorf("failed to connect to containerd after %d attempts: %w", i, err)
					}
				}

				log.Debug().Msg("importing images from embedded tarball")
				images, err := c.client.Import(c.ctx, bytes.NewReader(tar))
				if err != nil {
					return fmt.Errorf("failed to import task images (this indicates something wrong with the binary): %w", err)
				}

				for i, image := range images {
					log.Info().Msgf("imported image #%d '%s'", i, image.Name)
				}

				close(rdy)
				<-sgnls
				return nil
			},
		),
	)

	return component.NewNamedGroup(c.Name(), components...).Run(signals, ready)
}

func (c *ContainerdRuntime) Execute(_ []string) error {
	return <-ifrit.Invoke(c).Wait()
}

func (c *ContainerdRuntime) Name() string {
	return "containerd"
}

func (c *ContainerdRuntime) IsEnabled() bool {
	// at this point in time, we have no intention of writing
	// an alternative runtime implementation
	return true
}

func (c *ContainerdRuntime) Send(m geocloud.Message) error {
	k, v := "id", m.ID()
	log.Info().Str(k, v).Msg("processing message")

	log.Trace().Str(k, v).Msg("getting job from datastore")
	j, err := c.ds.GetJob(m)
	if err != nil {
		return err
	}

	stderr := new(bytes.Buffer)
	defer func() {
		j.EndTime = time.Now()
		jobErr := stderr.Bytes()
		if len(jobErr) > 0 {
			j.Err = fmt.Errorf("%s", jobErr)
			j.Status = geocloud.Error
		} else if err != nil {
			j.Err = err
			j.Status = geocloud.Error
		} else {
			j.Status = geocloud.Complete
		}
		log.Err(j.Err).Str(k, v).Msgf("job finished with status %s", j.Status.Status())
		c.ds.UpdateJob(j)
	}()

	j.Status = geocloud.InProgress
	log.Trace().Str(k, v).Msgf("setting job to %s", j.Status.Status())
	j, err = c.ds.UpdateJob(j)
	if err != nil {
		return err
	}

	log.Trace().Str(k, v).Msg("getting task for job from datastore")
	t, err := c.ds.GetTaskByJobID(m)
	if err != nil {
		return err
	}

	var (
		image  containerd.Image
		imgrdy = make(chan error, 1)
		volrdy = make(chan error, 1)
	)
	go func() {
		log.Debug().Str(k, v).Msgf("pulling image %s", t.Ref)
		if image, err = c.pull(t.Ref); err != nil {
			imgrdy <- err
		} else {
			close(imgrdy)
		}
	}()

	log.Trace().Str(k, v).Msg("creating input volume")
	invol, err := c.involume(m)
	if err != nil {
		return err
	}
	defer os.RemoveAll(c.jobdir(m))

	log.Trace().Str(k, v).Msg("getting input")
	input, err := c.os.GetInput(m)
	if err != nil {
		return err
	}

	go func() {
		log.Debug().Str(k, v).Msg("downloading input")
		if err = input.Download(invol.path); err != nil {
			volrdy <- err
		} else {
			close(volrdy)
		}
	}()

	log.Trace().Str(k, v).Msg("creating output volume")
	outvol, err := c.outvolume(m)
	if err != nil {
		return err
	}

	if err = <-imgrdy; err != nil {
		return err
	}
	log.Debug().Str(k, v).Msgf("finished pulling image %s", t.Ref)
	if err = <-volrdy; err != nil {
		return err
	}
	log.Debug().Str(k, v).Msg("finished downloading input")

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

	inmountdest, outmountdest := "/job/input", "/job/output"
	args := append([]string{filepath.Join(inmountdest, filename), outmountdest}, j.Args...)
	mounts := []specs.Mount{
		c.mount(invol.path, inmountdest, "rw"),
		c.mount(outvol.path, outmountdest, "rw"),
	}

	var f, a string
	for _, m := range mounts {
		f += fmt.Sprintf(" -v %s:%s", m.Source, m.Destination)
	}
	for _, r := range args {
		a += fmt.Sprintf(" %s", r)
	}

	log.Info().Str(k, v).Msgf("running%s %s%s", f, t.Ref, a)
	container, err := c.run(m, image, args, mounts)
	if err != nil {
		return err
	}
	defer container.Delete(c.ctx, containerd.WithSnapshotCleanup)

	log.Trace().Str(k, v).Msg("creating task")
	task, err := container.NewTask(c.ctx, cio.NewCreator(cio.WithStreams(os.Stdin, os.Stdout, stderr)))
	if err != nil {
		return err
	}
	defer task.Delete(c.ctx)

	log.Trace().Str(k, v).Msg("waiting on task to set up")
	exitStatusC, err := task.Wait(c.ctx)
	if err != nil {
		return err
	}

	log.Trace().Str(k, v).Msg("starting task")
	if err := task.Start(c.ctx); err != nil {
		return err
	}

	log.Trace().Str(k, v).Msg("waiting on task to complete")
	exitStatus := <-exitStatusC
	if exitCode, _, err := exitStatus.Result(); err != nil {
		return err
	} else if exitCode != 0 {
		err = fmt.Errorf("job exited with code %d", exitCode)
		return err
	}

	log.Debug().Str(k, v).Msg("uploading output")
	if err = c.os.PutOutput(m, outvol); err != nil {
		return err
	}

	return nil
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
	img, err := c.client.ImageService().Get(c.ctx, ref)
	if err != nil {
		img, err = c.client.Fetch(
			c.ctx, ref,
			containerd.WithResolver(*c.resolver),
		)
		if err != nil {
			return nil, err
		}
	}

	image := containerd.NewImage(c.client, img)
	return image, image.Unpack(c.ctx, containerd.DefaultSnapshotter)
}

func volume(path string) (*dirVolume, error) {
	return &dirVolume{path: path}, os.MkdirAll(path, 0755)
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
		Source:      src,
		Destination: dst,
		Type:        "bind",
		Options:     append([]string{"bind"}, opts...),
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
