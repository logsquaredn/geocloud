package migratecmd

import (
	"embed"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/logsquaredn/geocloud/shared/groups"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type MigrateCmd struct {
	Version  func() `long:"version" short:"v" description:"Print the version"`
	Loglevel string `long:"log-level" short:"l" default:"debug" choice:"trace" choice:"debug" choice:"info" choice:"warn" choice:"error" choice:"fatal" choice:"panic" description:"Geocloud log level"`

	groups.Postgres `group:"Postgres" namespace:"postgres"`
}

//go:embed migrations/*.up.sql
var migrations embed.FS

func (cmd *MigrateCmd) Execute(args []string) error {
	loglevel, err := zerolog.ParseLevel(cmd.Loglevel)
	if err != nil {
		log.Err(err).Msg("migrate exiting with error")
		return fmt.Errorf("migratecmd: failed to parse --log-level: %w", err)
	}
	zerolog.SetGlobalLevel(loglevel)

	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		log.Err(err).Msg("migrate exiting with error")
		return fmt.Errorf("migratecmd: failed to read migrations: %w", err)
	}

	var (
		m *migrate.Migrate
		i = 1
	)
	for m, err = migrate.NewWithSourceInstance(
		"migrations", source,
		cmd.Postgres.ConnectionString(),
	); err != nil; i++ {
		if i >= cmd.Postgres.Retries {
			log.Err(err).Msg("migrate exiting with error")
			return fmt.Errorf("migratecmd: failed to create migration after %d attempts: %w", i, err)
		}
		time.Sleep(cmd.Postgres.RetryDelay)
	}
	defer m.Close()

	return m.Up()
}
