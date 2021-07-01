package workercmd

import (
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/logsquaredn/geocloud/worker"
	"github.com/logsquaredn/geocloud/worker/listener"
	"github.com/rs/zerolog"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
)

type SQS struct {
	QueueNames  []string      `long:"queue-name" description:"SQS queue names to listen on"`
	QueueURLs   []string      `long:"queue-url" description:"SQS queue urls to listen on"`
	Visibility  time.Duration `long:"visibility-timeout" default:"1h" description:"Visibilty timeout for SQS messages"`
}

type AWS struct {
	AccessKeyID     string `long:"access-key-id" description:"AWS access key ID"`
	SecretAccessKey string `long:"secret-access-key" description:"AWS secret access key"`
	Region          string `long:"region" description:"AWS region"`

	SQS `group:"SQS" namespace:"sqs"`
}

type Registry struct {
	Address  string `long:"address" default:"registry-1.docker.io" description:"URL of the registry to pull images from"`
	Password string `long:"password" description:"Password to use to authenticate with the registry"`
	Username string `long:"username" description:"Username to use to authenticate with the registry"`
}

type WorkerCmd struct {
	Version    func() `short:"v" long:"version" description:"Print the version"`
	Loglevel   string `long:"log-level" short:"l" default:"debug" choice:"trace" choice:"debug" choice:"info" choice:"warn" choice:"error" choice:"fatal" choice:"panic" description:"Geocloud log level"`

	AWS        `group:"AWS" namespace:"aws"`
	Containerd `group:"Containerd" namespace:"containerd"`
	Registry   `group:"Registry" namespace:"registry"`
}

func (cmd *WorkerCmd) Execute(args []string) error {
	loglevel, err := zerolog.ParseLevel(cmd.Loglevel)
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(loglevel)

	var members grouper.Members

	if !cmd.Containerd.NoRun {
		containerd, err := cmd.containerd()
		if err != nil {
			return err
		}

		members = append(members, grouper.Member{
			Name: "containerd",
			Runner: containerd,
		})
	}

	http := http.DefaultClient

	listener, err := listener.New(
		listener.WithHttpClient(http),
		listener.WithRegion(cmd.Region),
		listener.WithQueueUrls(cmd.QueueURLs...),
		listener.WithQueueNames(cmd.QueueNames...),
		listener.WithVisibilityTimeout(cmd.Visibility),
		listener.WithCreds(credentials.NewStaticCredentials(cmd.AccessKeyID, cmd.SecretAccessKey, "")),
		listener.WithCallback(func(m worker.Message) error {
			return nil
		}),
	)
	if err != nil {
		return err
	}

	members = append(members, grouper.Member{
		Name: "listener",
		Runner: listener,
	})

	return <-ifrit.Invoke(grouper.NewOrdered(os.Interrupt, members)).Wait()
}
