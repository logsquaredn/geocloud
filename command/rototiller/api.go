package command

import (
	"fmt"
	"net"
	"net/http"

	"github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/api"
	"github.com/logsquaredn/rototiller/store/blob/bucket"
	"github.com/logsquaredn/rototiller/store/data/postgres"
	"github.com/logsquaredn/rototiller/stream/event/amqp"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func NewAPI() *cobra.Command {
	var (
		port                               int64
		postgresAddr, bucketAddr, amqpAddr string
		apiCmd                             = &cobra.Command{
			Use:     "api",
			Aliases: []string{"a"},
			RunE: func(cmd *cobra.Command, args []string) error {
				var (
					ctx  = cmd.Context()
					logr = rototiller.LoggerFrom(ctx)
				)

				migrations, err := postgres.NewMigrations(ctx, postgresAddr)
				if err != nil {
					return err
				}

				if err = migrations.Up(); err != nil {
					return err
				}

				datastore, err := postgres.New(ctx, postgresAddr)
				if err != nil {
					return err
				}

				eventStream, err := amqp.New(ctx, amqpAddr)
				if err != nil {
					return err
				}

				eventStreamProducer, err := eventStream.NewProducer(ctx)
				if err != nil {
					return err
				}

				blobstore, err := bucket.New(ctx, bucketAddr)
				if err != nil {
					return err
				}

				srv, err := api.NewHandler(ctx, datastore, eventStreamProducer, blobstore)
				if err != nil {
					return err
				}

				addr := fmt.Sprintf(":%d", port)
				l, err := net.Listen("tcp", addr)
				if err != nil {
					return err
				}

				logr.Info("serving on " + addr)
				return http.Serve(l, h2c.NewHandler(srv, &http2.Server{})) //nolint:gosec,nolintlint // lint in GitHub Actions doesn't like this
			},
		}
	)

	apiCmd.Flags().StringVar(&amqpAddr, "amqp-addr", "", "AMQP address")
	apiCmd.Flags().StringVar(&bucketAddr, "bucket-addr", "", "bucket address")
	apiCmd.Flags().StringVar(&postgresAddr, "postgres-addr", "", "Postgres address")
	apiCmd.Flags().Int64VarP(&port, "port", "p", 8080, "listen port")

	return apiCmd
}
