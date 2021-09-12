package aggregator

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/remotes"
	"github.com/logsquaredn/geocloud"
	"github.com/logsquaredn/geocloud/shared/das"
	"github.com/logsquaredn/geocloud/shared/oas"
	"github.com/logsquaredn/geocloud/worker"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/rs/zerolog/log"
)

type f = map[string]interface{}

type S3Aggregrator struct {
	das       *das.Das
	oas       *oas.Oas
	addr      string
	network   string
	listen    net.Listener
	server    *http.Server
	cclient   *containerd.Client
	resolver  *remotes.Resolver
	host      string
	sock      string
	namespace string
	prefetch  bool
	workdir   string
	tasks     []string
}

var _ worker.Aggregator = (*S3Aggregrator)(nil)

const runner = "S3Aggregator"

func New(das *das.Das, oas *oas.Oas, opts ...S3AggregatorOpt) (*S3Aggregrator, error) {
	if das == nil {
		return nil, fmt.Errorf("aggregator: nil das")
	}

	if oas == nil {
		return nil, fmt.Errorf("aggregator: nil oas")
	}

	a := &S3Aggregrator{}
	for _, opt := range opts {
		opt(a)
	}

	if a.listen == nil {
		var err error
		a.listen, err = a.listener()
		if err != nil {
			return nil, fmt.Errorf("aggregator: unable to create listener: %w", err)
		}
	}

	a.das = das
	a.oas = oas

	mux := http.NewServeMux()
	mux.Handle("/api/v1/run", a)
	a.server = &http.Server{
		Addr:    a.addr,
		Handler: mux,
	}

	return a, nil
}

func (a *S3Aggregrator) Aggregate(ctx context.Context, m geocloud.Message) error {
	// j, err := a.das.GetJobByJobID(m.ID())
	// if err != nil {
	// 	log.Err(err).Fields(f{ "runner": runner }).Msg("error getting job")
	// 	return err
	// }

	t, err := a.das.GetTaskByJobID(m.ID())
	if err != nil {
		log.Err(err).Fields(f{ "runner": runner }).Msg("error getting task")
		return err
	}

	ctx = namespaces.WithNamespace(ctx, a.namespace)

	// TODO pull, download, etc. in goroutines for speed
	log.Trace().Fields(f{ "runner": runner, "ref": t.Ref }).Msg("pulling ref")
	image, err := a.pull(ctx, t.Ref)
	if err != nil {
		log.Err(err).Fields(f{ "runner": runner, "ref": t.Ref }).Msg("error pulling ref")
		return err
	}

	tmpDir, err := os.MkdirTemp(a.workdir, "*")
	if err != nil {
		log.Err(err).Fields(f{ "runner": runner }).Msg("error creating temp dir")
		return err
	}
	// defer os.RemoveAll(tmpDir)

	inDir := filepath.Join(tmpDir, "input")
	err = os.MkdirAll(inDir, 0755)
	if err != nil {
		log.Err(err).Fields(f{ "runner": runner }).Msg("error creating input dir")
		return err
	}

	log.Trace().Fields(f{ "runner": runner }).Msg("downloading input")
	input, err := a.oas.DownloadJobInputToDir(m.ID(), inDir)
	if err != nil {
		log.Err(err).Fields(f{ "runner": runner }).Msg("error downloading input")
		return err
	}

	outDir := filepath.Join(tmpDir, "output")
	err = os.MkdirAll(outDir, 0755)
	if err != nil {
		log.Err(err).Fields(f{ "runner": runner }).Msg("error creating output dir")
		return err
	}

	inDest := filepath.Join("/job", "input", filepath.Base(input.Name()))
	outDest := filepath.Join("/job", "output")
	mounts := []specs.Mount{
		{
			Source: input.Name(),
			Destination: inDest,
			Type: "bind",
			Options: []string{ "bind", "ro" },
		},
		{
			Source: outDir,
			Destination: outDest,
			Type: "bind",
			Options: []string{ "bind", "rw" },
		},
	}
	args := append([]string { inDest, outDest }, "2") // TODO append arg(s) from postgres instead of 2
	log.Debug().Fields(f{ "runner": runner }).Msg("creating container")
	container, err := a.cclient.NewContainer(
		ctx, m.ID(),
		containerd.WithNewSnapshot(m.ID(), image),
		containerd.WithNewSpec(
			oci.WithImageConfigArgs(image, args), 
			oci.WithMounts(mounts),
		),
	)
	if err != nil {
		log.Err(err).Fields(f{ "runner": runner }).Msg("error creating container")
		return err
	}
	defer container.Delete(ctx, containerd.WithSnapshotCleanup)

	log.Trace().Fields(f{ "runner": runner }).Msg("creating task")
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio)) // TODO stderr to buffer to save for later
	if err != nil {
		log.Err(err).Fields(f{ "runner": runner }).Msg("error creating task")
		return err
	}
	defer task.Delete(ctx)

	log.Trace().Fields(f{ "runner": runner }).Msg("waiting on task")
	exitStatusC, err := task.Wait(ctx)
	if err != nil {
		log.Err(err).Fields(f{ "runner": runner }).Msg("error waiting on task")
		return err
	}

	log.Trace().Fields(f{ "runner": runner }).Msg("starting task")
	if err := task.Start(ctx); err != nil {
		log.Err(err).Fields(f{ "runner": runner }).Msg("error starting task")
		return err
	}

	log.Trace().Fields(f{ "runner": runner }).Msg("waiting on task to exit")
	exitStatus := <-exitStatusC
	exitCode, _, err := exitStatus.Result()
	if err != nil {
		log.Err(err).Fields(f{ "runner": runner }).Msg("error getting task result")
		return err
	}

	if exitCode != 0 {
		log.Err(err).Fields(f{ "runner": runner }).Msg("task exited with nonzero exit code")
		return err
	}

	// TODO walk outDir and upload outputs through oas

	return nil
}

func (a *S3Aggregrator) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	if a.cclient == nil {
		log.Info().Fields(f{ "runner": runner }).Msg("creating containerd client")
		if a.namespace == "" {
			a.namespace = "geocloud"
		}

		var err error
		a.cclient, err = containerd.New(a.sock, containerd.WithDefaultNamespace(a.namespace))
		if err != nil {
			log.Err(err).Fields(f{ "runner": runner }).Msg("error creating containerd client")
			return fmt.Errorf("aggregator: unable to create containerd client: %w", err)
		}
	}

	ctx := namespaces.WithNamespace(context.Background(), a.namespace)

	if a.prefetch {
		log.Info().Fields(f{ "runner": runner }).Msg("getting task refs")
		refs, err := a.das.GetTaskRefsByTaskTypes(a.tasks...)
		if err != nil {
			log.Err(err).Fields(f{ "runner": runner }).Msg("error getting task refs")
			return fmt.Errorf("aggregator: unable to get task refs: %w", err)
		}
	
		log.Info().Fields(f{ "runner": runner, "refs": refs }).Msg("pulling task images")
		for _, ref := range refs {
			go a.pull(ctx, ref)
		}
	}

	wait := make(chan error, 1)
	go func() {
		wait<- a.server.Serve(a.listen)
	}()

	log.Info().Fields(f{ "runner": runner }).Msg("ready")
	close(ready)
	for {
		select {
		case err := <-wait:
			log.Err(err).Fields(f{ "runner": runner }).Msg("received error")
			defer a.server.Close()
			return fmt.Errorf("aggregator: received error: %w", err)
		case signal := <-signals:
			log.Info().Fields(f{ "runner": runner, "signal": signal.String() }).Msg("received signal")
			return a.server.Shutdown(ctx)
		}
	}
}

const (
	addr = "127.0.0.1:7777"
	tcp  = "tcp"
)

func (a *S3Aggregrator) listener() (net.Listener, error) {
	if a.addr == "" {
		a.addr = addr
	}

	if a.network == "" {
		a.network = tcp
	}

	return net.Listen(a.network, a.addr)
}
