package main

import (
	"github.com/logsquaredn/geocloud/datastore"
	"github.com/spf13/cobra"
)

var externalMigrateCmd = &cobra.Command{
	Use:     "external",
	Aliases: []string{"e"},
	RunE:    runExternalMigrate,
}

func runExternalMigrate(cmd *cobra.Command, args []string) error {
	ds, err := datastore.NewPostgres(
		getPostgresOpts(),
	)
	if err != nil {
		return err
	}

	return ds.MigrateExternal()
}
