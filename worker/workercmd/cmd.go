package workercmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/jessevdk/go-flags"
	"github.com/logsquaredn/geocloud/shared/das"
	"github.com/logsquaredn/geocloud/shared/groups"
	"github.com/logsquaredn/geocloud/shared/oas"
	"github.com/logsquaredn/geocloud/worker/aggregator"
	"github.com/logsquaredn/geocloud/worker/listener"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
)

type Registry struct {
	URL      string `long:"url" default:"https://registry-1.docker.io/v2/" description:"URL of the registry to pull images from"`
	Username string `long:"username" env:"GEOCLOUD_REGISTRY_USERNAME" description:"Username to use to authenticate with the registry"`
	Password string `long:"password" env:"GEOCLOUD_REGISTRY_PASSWORD" description:"Password to use to authenticate with the registry"`
}

type WorkerCmd struct {
	Version    func()         `long:"version" short:"v" description:"Print the version"`
	Loglevel   string         `long:"log-level" short:"l" default:"debug" choice:"trace" choice:"debug" choice:"info" choice:"warn" choice:"error" choice:"fatal" choice:"panic" description:"Geocloud log level"`
	IP         string         `long:"ip" default:"127.0.0.1" env:"GEOCLOUD_WORKER_IP" description:"IP for the worker to listen on"`
	Port       int64          `long:"port" default:"7778" description:"Port for the worker to listen on"`
	Tasks      []string       `long:"task" short:"t" description:"Task types that the worker should execute"`
    Workdir    flags.Filename `long:"workdir" short:"w" default:"/var/lib/geocloud" description:"Working directory for temporary files"`
	
	groups.AWS      `group:"AWS" namespace:"aws"`
	Containerd      `group:"Containerd" namespace:"containerd"`
	groups.Postgres `group:"Postgres" namespace:"postgres"`
	Registry        `group:"Registry" namespace:"registry"`
}

func (cmd *WorkerCmd) Execute(args []string) error {
	loglevel, err := zerolog.ParseLevel(cmd.Loglevel)
	if err != nil {
		log.Err(err).Msg("worker exiting with error")
		return fmt.Errorf("workercmd: failed to parse --log-level: %w", err)
	}
	zerolog.SetGlobalLevel(loglevel)

	workdir := string(cmd.Workdir)
	if err := os.MkdirAll(filepath.Dir(workdir), 0755); err != nil {
		return err
	}
	
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

	url, err := url.Parse(cmd.Registry.URL)
	if err != nil {
		return err
	}

	resolver := docker.NewResolver(docker.ResolverOptions{
		Hosts: func(s string) ([]docker.RegistryHost, error) {
			return []docker.RegistryHost{
				{
					Client: http,
					Authorizer: docker.NewDockerAuthorizer(
						docker.WithAuthCreds(func(s string) (string, string, error) {
							return cmd.Registry.Username, cmd.Registry.Password, nil
						}),
					),
					Scheme: url.Scheme,
					Host: url.Host,
					Path: url.Path,
					Capabilities: docker.HostCapabilityPull|docker.HostCapabilityResolve|docker.HostCapabilityPush,
				},
			}, nil
		},
	})
	ag, err := aggregator.New(
		da, oa,
		aggregator.WithAddress(fmt.Sprintf("%s:%d", cmd.IP, cmd.Port)),
		aggregator.WithContainerdNamespace(cmd.Containerd.Namespace),
		aggregator.WithContainerdSocket(string(cmd.Containerd.Address)),
		aggregator.WithPrefetch(true),
		aggregator.WithResolver(&resolver),
		aggregator.WithRegistryHost(url.Host),
		aggregator.WithTasks(cmd.Tasks...),
		aggregator.WithWorkdir(workdir),
	)
	if err != nil {
		log.Err(err).Msg("worker exiting with error")
		return fmt.Errorf("workercmd: failed to create aggregator: %w", err)
	}

	ln, err := listener.New(
		sess, ag.Aggregate, da,
		listener.WithTasks(cmd.Tasks...),
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
