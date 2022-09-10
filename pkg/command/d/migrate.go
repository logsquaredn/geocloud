package command

import (
	"github.com/logsquaredn/rototiller/pkg/store/data/postgres"
	"github.com/spf13/cobra"

	_ "gocloud.dev/blob/s3blob"
)

func NewMigrate() *cobra.Command {
	var (
		postgresAddr string
		migrateCmd   = &cobra.Command{
			Use:     "migrate",
			Aliases: []string{"m"},
			RunE: func(cmd *cobra.Command, args []string) error {
				var (
					ctx = cmd.Context()
				)

				migrations, err := postgres.NewMigrations(ctx, postgresAddr)
				if err != nil {
					return err
				}

				return migrations.Up()
			},
		}
	)

	migrateCmd.Flags().StringVar(&postgresAddr, "postgres-addr", "", "Postgres address")

	return migrateCmd
}
