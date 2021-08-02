package apicmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/logsquaredn/geocloud/api/janitor"
	"github.com/logsquaredn/geocloud/api/router"
	"github.com/logsquaredn/geocloud/shared/das"
	"github.com/logsquaredn/geocloud/shared/oas"
	"github.com/logsquaredn/geocloud/shared/groups"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
)

type APICmd struct {
	Version  func() `long:"version" short:"v" description:"Print the version"`
	Loglevel string `long:"log-level" short:"l" default:"debug" choice:"trace" choice:"debug" choice:"info" choice:"warn" choice:"error" choice:"fatal" choice:"panic" description:"Geocloud log level"`

	groups.AWS      `group:"AWS" namespace:"aws"`
	groups.Postgres `group:"Postgres" namespace:"postgres"`
}

func (cmd *APICmd) Execute(args []string) error {
	loglevel, err := zerolog.ParseLevel(cmd.Loglevel)
	if err != nil {
		log.Err(err).Msg("api exiting with error")
		return fmt.Errorf("apicmd: failed to parse --log-level: %w", err)
	}
	zerolog.SetGlobalLevel(loglevel)

	var members grouper.Members

	http := http.DefaultClient
	cfg := aws.NewConfig().WithHTTPClient(http).WithRegion(cmd.Region).WithCredentials(cmd.AWS.Credentials())
	sess, err := session.NewSession(cfg)
	if err != nil {
		log.Err(err).Msg("api exiting with error")
		return fmt.Errorf("apicmd: failed to create session: %w", err)
	}

	da, err := das.New(cmd.Postgres.ConnectionString(), das.WithRetries(cmd.Postgres.Retries), das.WithRetryDelay(cmd.Postgres.RetryDelay))
	if err != nil {
		log.Err(err).Msg("api exiting with error")
		return fmt.Errorf("apicmd: failed to create das: %w", err)
	}
	defer da.Close()

	oa, err := oas.New(sess, cmd.AWS.S3.Bucket, oas.WithPrefix(cmd.AWS.S3.Prefix))
	if err != nil {
		log.Err(err).Msg("api exiting with error")
		return fmt.Errorf("apicmd: failed to create oas: %w", err)
	}

	sqs := sqs.New(sess)

	rtr, err := router.New(da, oa, sqs)
	if err != nil {
		log.Err(err).Msg("api exiting with error")
		return fmt.Errorf("apicmd: failed to create router: %w", err)
	}

	members = append(members, grouper.Member{
		Name:   "router",
		Runner: rtr,
	})

	jn, err := janitor.New(da)
	if err != nil {
		log.Err(err).Msg("api exiting with error")
		return fmt.Errorf("apicmd: failed to create janitor: %w", err)
	}

	members = append(members, grouper.Member{
		Name:   "janitor",
		Runner: jn,
	})

	return <-ifrit.Invoke(grouper.NewOrdered(os.Interrupt, members)).Wait()
}
