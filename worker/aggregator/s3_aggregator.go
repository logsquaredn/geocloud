package aggregator

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/remotes"
	"github.com/logsquaredn/geocloud"
	"github.com/logsquaredn/geocloud/shared/das"
	"github.com/logsquaredn/geocloud/shared/oas"
	"github.com/logsquaredn/geocloud/worker"
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

func (a *S3Aggregrator) Aggregate(m geocloud.Message) error {
	task, err := a.das.GetTaskByJobID(m.ID())
	if err != nil {
		log.Err(err).Fields(f{ "runner": runner }).Msg("error getting task ref")
		return err
	}

	// TODO pull, download, etc. in goroutines for speed
	log.Trace().Fields(f{ "runner": runner, "ref": task.Ref }).Msg("pulling ref")
	_, err = a.pull(task.Ref)
	if err != nil {
		log.Err(err).Fields(f{ "runner": runner, "ref": task.Ref }).Msg("error pulling ref")
		return err
	}

	return nil
}

var ctx = context.Background()

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

	if a.prefetch {
		log.Info().Fields(f{ "runner": runner }).Msg("getting task refs")
		refs, err := a.das.GetTaskRefsByTaskTypes(a.tasks...)
		if err != nil {
			log.Err(err).Fields(f{ "runner": runner }).Msg("error getting task refs")
			return fmt.Errorf("aggregator: unable to get task refs: %w", err)
		}

		log.Info().Fields(f{ "runner": runner, "refs": refs }).Msg("pulling task images")
		for _, ref := range refs {
			go a.pull(ref)
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
