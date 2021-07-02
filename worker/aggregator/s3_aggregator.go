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
	"github.com/logsquaredn/geocloud/shared/das"
	"github.com/logsquaredn/geocloud/worker"
	"github.com/rs/zerolog/log"
)

type f = map[string]interface{}

type S3Aggregrator struct {
	svc       *s3.S3
	sess      *session.Session
	creds     *credentials.Credentials
	region    string
	bucket    string
	prefix    string
	addr      string
	network   string
	hclient   *http.Client
	conn      string
	db        *sql.DB
	das       *das.Das
	listen    net.Listener
	server    *http.Server
	cclient   *containerd.Client
	sock      string
	namespace string
}

var _ worker.Aggregator = (*S3Aggregrator)(nil)

const runner = "S3Aggregator"

func New(opts ...S3AggregatorOpt) (*S3Aggregrator, error) {
	a := &S3Aggregrator{}
	for _, opt := range opts {
		opt(a)
	}

	if a.das == nil {
		var err error
		a.das, err = das.New(das.WithConnectionString(a.conn))
		if err != nil {
			return nil, err
		}
	}

	if a.hclient == nil {
		a.hclient = http.DefaultClient
	}

	if a.svc == nil {
		var err error
		a.svc, err = a.service()
		if err != nil {
			return nil, err
		}
	}

	if a.listen == nil {
		var err error
		a.listen, err = a.listener()
		if err != nil {
			return nil, err
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/api/v1/run", &s3Handler{
		bucket: a.bucket,
		das: a.das,
		prefix: a.prefix,
		svc: a.svc,
	})
	a.server = &http.Server{
		Addr: a.addr,
		Handler: mux,
	}

	return a, nil
}

func (a *S3Aggregrator) Aggregate(m worker.Message) error {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/run?id=%s", a.server.Addr, m.ID()), nil)
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
		wait<- a.server.Serve(a.listen)
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

const (
	addr   = "127.0.0.1:7777"
	tcp    = "tcp"
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

func (a *S3Aggregrator) service() (*s3.S3, error) {
	if a.sess == nil {
		var err error
		a.sess, err = a.session()
		if err != nil {
			return nil, err
		}
	}

	return s3.New(a.sess), nil
}

func (a *S3Aggregrator) session() (*session.Session, error) {
	if a.hclient == nil {
		a.hclient = http.DefaultClient
	}

	cfg := aws.NewConfig().WithHTTPClient(a.hclient).WithRegion(a.region)
	if a.creds != nil {
		cfg = cfg.WithCredentials(a.creds)
	} else {
		cfg = cfg.WithCredentials(credentials.NewEnvCredentials())
	}

	return session.NewSession(cfg)
}
