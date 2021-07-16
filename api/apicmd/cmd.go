package apicmd

import (
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/logsquaredn/geocloud/api/janitor"
	"github.com/logsquaredn/geocloud/api/router"
	"github.com/logsquaredn/geocloud/shared/das"
	"github.com/logsquaredn/geocloud/shared/oas"
	"github.com/rs/zerolog/log"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
)

type Postgres struct {
	Host     string `long:"host" description:"Postgres host"`
	Port     int    `long:"port" default:"5432" description:"Postgres port"`
	User     string `long:"user" default:"geocloud" description:"Postgres username"`
	Password string `long:"password" description:"Postgres password"`
	SSLMode  string `long:"ssl-mode" default:"disable" choice:"disable" description:"Postgres SSL mode"`
	Retries  int    `long:"retries" default:"5" description:"Number of times to retry connecting to Postgres"`
}

type APICmd struct {
	Version func() `short:"v" long:"version" description:"Print the version"`

	Postgres `group:"Postgres" namespace:"postgres"`
}

func (cmd *APICmd) Execute(args []string) error {
	var members grouper.Members

	var (
		da    *das.Das
		daErr error
	)
	if da, daErr = das.New(das.WithConnectionString(cmd.getConnectionString()), das.WithRetries(cmd.Postgres.Retries)); daErr != nil {
		return fmt.Errorf("apicmd: failed to create das: %w", daErr)
	}
	defer da.Close()

	var (
		oa    *oas.Oas
		oaErr error
	)
	if oa, oaErr = oas.New(oas.WithBucket(""), oas.WithRegion("us-east-1")); oaErr != nil {
		return fmt.Errorf("apicmd: failed to create oas: %w", oaErr)
	}

	rtr, err := router.New(router.WithDas(da), router.WithOas(oa))
	if err != nil {
		log.Err(err).Msg("failed to create router")
		return err
	}

	members = append(members, grouper.Member{
		Name:   "router",
		Runner: rtr,
	})

	jn, err := janitor.New(janitor.WithDas(da))
	if err != nil {
		log.Err(err).Msg("failed to create jantior")
		return err
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
