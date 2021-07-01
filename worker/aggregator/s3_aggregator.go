package aggregator

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/containerd/containerd"
	"github.com/logsquaredn/geocloud/worker"
	"github.com/rs/zerolog/log"
)

type f = map[string]interface{}

type S3Aggregrator struct {
	svc       *s3.S3
	sess      *session.Session
	creds     *credentials.Credentials
	region    string
	addr      string
	network   string
	hclient   *http.Client
	db        *sql.DB
	listener  net.Listener
	server    *http.Server
	cclient   *containerd.Client
	sock      string
	namespace string
}

var _ worker.Aggregator = (*S3Aggregrator)(nil)

const (
	addr   = "127.0.0.1:7777"
	tcp    = "tcp"
	runner = "S3Aggregator"
)

func New(opts ...S3AggregatorOpt) (*S3Aggregrator, error) {
	a := &S3Aggregrator{}
	for _, opt := range opts {
		opt(a)
	}

	if a.db == nil {
		return nil, fmt.Errorf("nil db")
	}

	if a.hclient == nil {
		a.hclient = http.DefaultClient
	}

	if a.svc == nil {
		if a.sess == nil {
			var err error
			a.sess, err = session.NewSession(
				aws.NewConfig().WithCredentials(a.creds).WithHTTPClient(a.hclient).WithRegion(a.region),
			)
			if err != nil {
				return nil, err
			}
		}

		a.svc = s3.New(a.sess)
	}

	if a.listener == nil {
		if a.network == "" {
			a.network = tcp
		}

		if a.addr == "" {
			a.addr = addr
		}

		var err error
		a.listener, err = net.Listen(a.network, a.addr)
		if err != nil {
			return nil, err
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/api/v1/run", &s3AggregatorHandler{
		db:  a.db,
		svc: a.svc,
	})
	a.server = &http.Server{
		Addr: a.addr,
		Handler: mux,
	}

	return a, nil
}

func (a *S3Aggregrator) Aggregate(m worker.Message) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/run?id=%s", a.server.Addr, m.ID()), nil)
	if err != nil {
		return err
	}

	resp, err := a.hclient.Do(req)
	if err != nil {
		return err
	}

	return resp.Body.Close()
}

var ctx = context.Background()

func (a *S3Aggregrator) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	if a.cclient == nil {
		if a.namespace == "" {
			a.namespace = "geocloud"
		}

		var err error
		a.cclient, err = containerd.New(a.sock, containerd.WithDefaultNamespace(a.namespace))
		if err != nil {
			return err
		}
	}

	wait := make(chan error, 1)
	go func() {
		wait<- a.server.Serve(a.listener)
	}()

	log.Debug().Fields(f{ "runner":runner }).Msg("ready")
	close(ready)
	for {
		select {
		case err := <-wait:
			log.Error().Err(err).Fields(f{ "runner":runner }).Msg("received error")
			defer a.server.Close()
			return err
		case signal := <-signals:
			log.Debug().Fields(f{ "runner":runner, "signal":signal.String() }).Msg("received signal")
			return a.server.Shutdown(ctx)
		}
	}
}
