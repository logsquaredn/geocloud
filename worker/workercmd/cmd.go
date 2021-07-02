package workercmd

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	_ "github.com/lib/pq"
	"github.com/logsquaredn/geocloud/worker/aggregator"
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

type Postgres struct {
	Host     string `long:"host" description:"Postgres host"`
	Port     int64  `long:"port" default:"5432" description:"Postgres port"`
	User     string `long:"user" default:"geocloud" description:"Postgres user"`
	Password string `long:"password" description:"Postgres password"`
	SSLMode  string `long:"ssl-mode" default:"disable" choice:"disable" description:"Postgres SSL mode"`
}

type Registry struct {
	Address  string `long:"address" default:"registry-1.docker.io" description:"URL of the registry to pull images from"`
	Password string `long:"password" description:"Password to use to authenticate with the registry"`
	Username string `long:"username" description:"Username to use to authenticate with the registry"`
}

type WorkerCmd struct {
	Version    func() `short:"v" long:"version" description:"Print the version"`
	Loglevel   string `long:"log-level" short:"l" default:"debug" choice:"trace" choice:"debug" choice:"info" choice:"warn" choice:"error" choice:"fatal" choice:"panic" description:"Geocloud log level"`
	IP         string `long:"ip" default:"127.0.0.1" description:"IP for the worker to listen on"`
	Port       int64  `long:"port" default:"7777" description:"Port for the worker to listen on"`

	AWS        `group:"AWS" namespace:"aws"`
	Containerd `group:"Containerd" namespace:"containerd"`
	Postgres   `group:"Postgres" namespace:"postgres"`
	Registry   `group:"Registry" namespace:"registry"`
}

const (
	driver = "postgres"
)

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
	sess, err := session.NewSession(
		aws.NewConfig().WithCredentials(
			credentials.NewStaticCredentials(cmd.AccessKeyID, cmd.SecretAccessKey, ""),
		).WithHTTPClient(http).WithRegion(cmd.Region),
	)
	if err != nil {
		return err
	}

	var db *sql.DB
	for db, err = sql.Open(driver, cmd.getConnectionString()); err != nil; {
		time.Sleep(time.Second*15)
	}
	// db.SetMaxOpenConns

	ag, err := aggregator.New(
		aggregator.WithDB(db),
		aggregator.WithSession(sess),
		aggregator.WithRegion(cmd.Region),
		aggregator.WithHttpClient(http),
		aggregator.WithAddress(fmt.Sprintf("%s:%d", cmd.IP, cmd.Port)),
		aggregator.WithContainerdSocket(string(cmd.Containerd.Address)),
	)
	if err != nil {
		return err
	}

	ln, err := listener.New(
		listener.WithSession(sess),
		listener.WithRegion(cmd.Region),
		listener.WithQueueUrls(cmd.QueueURLs...),
		listener.WithQueueNames(cmd.QueueNames...),
		listener.WithVisibilityTimeout(cmd.Visibility),
		listener.WithCallback(ag.Aggregate),
	)
	if err != nil {
		return err
	}

	members = append(members, grouper.Member{
		Name: "listener",
		Runner: ln,
	}, grouper.Member{
		Name: "aggregator",
		Runner: ag,
	})

	return <-ifrit.Invoke(grouper.NewOrdered(os.Interrupt, members)).Wait()
}

func (cmd *WorkerCmd) getConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d?sslmode=%s", cmd.Postgres.User, cmd.Postgres.Password, cmd.Postgres.Host, cmd.Postgres.Port, cmd.Postgres.SSLMode)
}
