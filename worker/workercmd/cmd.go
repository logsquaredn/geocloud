package workercmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/logsquaredn/geocloud/shared/das"
	"github.com/logsquaredn/geocloud/shared/oas"
	"github.com/logsquaredn/geocloud/shared/sharedcmd"
	"github.com/logsquaredn/geocloud/worker/aggregator"
	"github.com/logsquaredn/geocloud/worker/listener"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
)

type Registry struct {
	Address  string `long:"address" default:"registry-1.docker.io" description:"URL of the registry to pull images from"`
	Password string `long:"password" description:"Password to use to authenticate with the registry"`
	Username string `long:"username" description:"Username to use to authenticate with the registry"`
}

type WorkerCmd struct {
	Version    func() `long:"version" short:"v" description:"Print the version"`
	Loglevel   string `long:"log-level" short:"l" default:"debug" choice:"trace" choice:"debug" choice:"info" choice:"warn" choice:"error" choice:"fatal" choice:"panic" description:"Geocloud log level"`
	IP         string `long:"ip" default:"127.0.0.1" env:"GEOCLOUD_WORKER_IP" description:"IP for the worker to listen on"`
	Port       int64  `long:"port" default:"7778" description:"Port for the worker to listen on"`

	sharedcmd.AWS      `group:"AWS" namespace:"aws"`
	Containerd         `group:"Containerd" namespace:"containerd"`
	sharedcmd.Postgres `group:"Postgres" namespace:"postgres"`
	Registry           `group:"Registry" namespace:"registry"`
}

func (cmd *WorkerCmd) Execute(args []string) error {
	loglevel, err := zerolog.ParseLevel(cmd.Loglevel)
	if err != nil {
		log.Err(err).Msg("worker exiting with error")
		return fmt.Errorf("workercmd: failed to parse --log-level: %w", err)
	}
	zerolog.SetGlobalLevel(loglevel)

	var members grouper.Members

	if !cmd.Containerd.NoRun {
		containerd, err := cmd.containerd()
		if err != nil {
			log.Err(err).Msg("worker exiting with error")
			return fmt.Errorf("workercmd: failed to execute containerd: %w", err)
		}

		members = append(members, grouper.Member{
			Name: "containerd",
			Runner: containerd,
		})
	}

	http := http.DefaultClient
	cfg := aws.NewConfig().WithHTTPClient(http).WithRegion(cmd.Region).WithCredentials(cmd.AWS.Credentials())
	sess, err := session.NewSession(cfg)
	if err != nil {
		log.Err(err).Msg("worker exiting with error")
		return fmt.Errorf("workercmd: failed to create session: %w", err)
	}

	da, err := das.New(cmd.Postgres.ConnectionString(), das.WithRetries(cmd.Postgres.Retries), das.WithRetryDelay(cmd.Postgres.RetryDelay))
	if err != nil {
		log.Err(err).Msg("worker exiting with error")
		return fmt.Errorf("workercmd: failed to create das: %w", err)
	}
	defer da.Close()

	oa, err := oas.New(sess, cmd.AWS.S3.Bucket)
	if err != nil {
		log.Err(err).Msg("worker exiting with error")
		return fmt.Errorf("workercmd: failed to create oas: %w", err)
	}

	ag, err := aggregator.New(
		da, oa,
		aggregator.WithHttpClient(http),
		aggregator.WithAddress(fmt.Sprintf("%s:%d", cmd.IP, cmd.Port)),
		aggregator.WithContainerdNamespace(cmd.Containerd.Namespace),
		aggregator.WithContainerdSocket(string(cmd.Containerd.Address)),
	)
	if err != nil {
		log.Err(err).Msg("worker exiting with error")
		return fmt.Errorf("workercmd: failed to create aggregator: %w", err)
	}

	ln, err := listener.New(
		sess, ag.Aggregate,
		listener.WithQueueUrls(cmd.AWS.SQS.QueueURLs...),
		listener.WithQueueNames(cmd.AWS.SQS.QueueNames...),
		listener.WithVisibilityTimeout(cmd.AWS.SQS.Visibility),
	)
	if err != nil {
		log.Err(err).Msg("worker exiting with error")
		return fmt.Errorf("workercmd: failed to create listener: %w", err)
	}

	members = append(members, grouper.Member{
		Name: "aggregator",
		Runner: ag,
	}, grouper.Member{
		Name: "listener",
		Runner: ln,
	})

	return <-ifrit.Invoke(grouper.NewOrdered(os.Interrupt, members)).Wait()
}
