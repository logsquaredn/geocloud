package main

import (
	"github.com/logsquaredn/rototiller/datastore"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:     "migrate",
	Aliases: []string{"m"},
	RunE:    runMigrate,
}

func runMigrate(cmd *cobra.Command, args []string) error {
	m, err := datastore.NewPostgresMigrations(
		getPostgresOpts(),
	)
	if err != nil {
		return err
	}

	log.Info().Msg("running migrations")
	return m.Migrate()
}
