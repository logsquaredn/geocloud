package main

import (
	"github.com/logsquaredn/geocloud/datastore"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:     "migrate",
	Aliases: []string{"m"},
	RunE:    runMigrate,
}

func init() {
	migrateCmd.AddCommand(
		coreMigrateCmd,
		externalMigrateCmd,
	)
}

func runMigrate(cmd *cobra.Command, args []string) error {
	ds, err := datastore.NewPostgres(
		getPostgresOpts(),
	)
	if err != nil {
		return err
	}

	if err := ds.MigrateCore(); err != nil {
		return err
	}

	return ds.MigrateExternal()
}
