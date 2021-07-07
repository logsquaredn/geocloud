package apicmd

import (
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/logsquaredn/geocloud/api/janitor"
	"github.com/logsquaredn/geocloud/api/router"
	"github.com/logsquaredn/geocloud/shared/das"
	"github.com/rs/zerolog/log"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
)

type Postgres struct {
	Host     string `long:"host" description:"Postgres host"`
	Port     int    `long:"port" default:"5432" description:"Postgres port"`
	User     string `long:"user" default:"geocloud" description:"Postgres username"`
	Password string `long:"password" description:"Postgres password"`
}

type APICmd struct {
	Version func() `short:"v" long:"version" description:"Print the version"`

	Postgres `group:"Postgres" namespace:"postgres"`
}

func (cmd *APICmd) Execute(args []string) error {
	var members grouper.Members

	var (
		da  *das.Das
		err error
	)
	if da, err = das.New(das.WithConnectionString(cmd.getDBConnectionString())); err != nil {
		return fmt.Errorf("apicmd: failed to create das: %w", err)
	}
	defer da.Close()

	rtr, err := router.New(router.WithDas(da))
	if err != nil {
		log.Err(err).Msg("failed to create router")
		return err
	}

	members = append(members, grouper.Member{
		Name: "router",
		Runner: rtr,
	})

	jn, err := janitor.New(janitor.WithDas(da))
	if err != nil {
		log.Err(err).Msg("failed to create jantior")
		return err
	}

	members = append(members, grouper.Member{
		Name: "janitor",
		Runner: jn,
	})
	
	return <-ifrit.Invoke(grouper.NewOrdered(os.Interrupt, members)).Wait()
}

func (cmd *APICmd) getDBConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d?sslmode=disable", cmd.Postgres.User, cmd.Postgres.Password, cmd.Postgres.Host, cmd.Postgres.Port)
}
