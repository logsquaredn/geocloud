package apicmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jessevdk/go-flags"
	"github.com/logsquaredn/geocloud/api/janitor"
	"github.com/logsquaredn/geocloud/api/router"
	"github.com/logsquaredn/geocloud/shared/das"
	"github.com/logsquaredn/geocloud/shared/oas"
	"github.com/rs/zerolog"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
)

type S3 struct {
	Bucket string `long:"bucket" description:"S3 bucket"`
	Prefix string `long:"prefix" default:"jobs" description:"S3 key prefix"`
}

type SQS struct {
	QueueNames  []string `long:"queue-name" description:"SQS queue names to listen on"`
	QueueURLs   []string `long:"queue-url" description:"SQS queue urls to listen on"`
}

type AWS struct {
	AccessKeyID     string         `long:"access-key-id" description:"AWS access key ID"`
	SecretAccessKey string         `long:"secret-access-key" description:"AWS secret access key"`
	Region          string         `long:"region" default:"us-east-1" description:"AWS region"`
	Profile         string         `long:"profile" description:"AWS profile"`
	SharedCreds     flags.Filename `long:"shared-credentials-file" description:"Path to AWS shared credentials file"`

	S3  `group:"S3" namespace:"s3"`
	SQS `group:"SQS" namespace:"sqs"`
}

type Postgres struct {
	Host     string `long:"host" description:"Postgres host"`
	Port     int    `long:"port" default:"5432" description:"Postgres port"`
	User     string `long:"user" default:"geocloud" description:"Postgres username"`
	Password string `long:"password" description:"Postgres password"`
	SSLMode  string `long:"ssl-mode" default:"disable" choice:"disable" description:"Postgres SSL mode"`
	Retries  int    `long:"retries" default:"5" description:"Number of times to retry connecting to Postgres"`
}

type APICmd struct {
	Version func() `long:"version" short:"v" description:"Print the version"`
	Loglevel   string `long:"log-level" short:"l" default:"debug" choice:"trace" choice:"debug" choice:"info" choice:"warn" choice:"error" choice:"fatal" choice:"panic" description:"Geocloud log level"`

	AWS      `group:"AWS" namespace:"aws"`
	Postgres `group:"Postgres" namespace:"postgres"`
}

func (cmd *APICmd) Execute(args []string) error {
	loglevel, err := zerolog.ParseLevel(cmd.Loglevel)
	if err != nil {
		return fmt.Errorf("apicmd: failed to parse log-level: %w", err)
	}
	zerolog.SetGlobalLevel(loglevel)

	var members grouper.Members

	http := http.DefaultClient
	cfg := aws.NewConfig().WithHTTPClient(http).WithRegion(cmd.Region).WithCredentials(
		credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.StaticProvider{
					Value: credentials.Value{
						AccessKeyID: cmd.AccessKeyID,
						SecretAccessKey: cmd.SecretAccessKey,
					},
				},
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{
					Filename: string(cmd.SharedCreds),
					Profile: cmd.Profile,
				},
			},
		),
	)
	sess, err := session.NewSession(cfg)
	if err != nil {
		return fmt.Errorf("apicmd: failed to create session: %w", err)
	}

	da, err := das.New(cmd.getConnectionString(), das.WithRetries(cmd.Postgres.Retries))
	if err != nil {
		return fmt.Errorf("apicmd: failed to create das: %w", err)
	}
	defer da.Close()

	oa, err := oas.New(sess, cmd.AWS.S3.Bucket, oas.WithPrefix(cmd.AWS.S3.Prefix))
	if err != nil {
		return fmt.Errorf("apicmd: failed to create oas: %w", err)
	}

	rtr, err := router.New(da, oa)
	if err != nil {
		return fmt.Errorf("apicmd: failed to create router: %w", err)
	}

	members = append(members, grouper.Member{
		Name:   "router",
		Runner: rtr,
	})

	jn, err := janitor.New(da)
	if err != nil {
		return fmt.Errorf("apicmd: failed to create janitor: %w", err)
	}

	members = append(members, grouper.Member{
		Name:   "janitor",
		Runner: jn,
	})

	return <-ifrit.Invoke(grouper.NewOrdered(os.Interrupt, members)).Wait()
}

func (cmd *APICmd) getConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d?sslmode=%s", cmd.Postgres.User, cmd.Postgres.Password, cmd.Postgres.Host, cmd.Postgres.Port, cmd.Postgres.SSLMode)
}
