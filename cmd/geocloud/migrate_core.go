package main

import (
	"github.com/logsquaredn/geocloud/datastore"
	"github.com/spf13/cobra"
)

var coreMigrateCmd = &cobra.Command{
	Use:     "core",
	Aliases: []string{"c"},
	RunE:    runCoreMigrate,
}

func runCoreMigrate(cmd *cobra.Command, args []string) error {
	ds, err := datastore.NewPostgres(
		getPostgresOpts(),
	)
	if err != nil {
		return err
	}

	return ds.MigrateCore()
}
