package command

import (
	"github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/pkg/api"
	"github.com/logsquaredn/rototiller/pkg/store/blob/bucket"
	"github.com/logsquaredn/rototiller/pkg/store/data/postgres"
	"github.com/logsquaredn/rototiller/pkg/stream/event/amqp"
	"github.com/logsquaredn/rototiller/pkg/worker"
	"github.com/spf13/cobra"
)

func NewWorker() *cobra.Command {
	var (
		workingDir, postgresAddr, bucketAddr, amqpAddr string
		workerCmd                                      = &cobra.Command{
			Use:     "worker",
			Aliases: []string{"w"},
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

				eventStreamConsumer, err := eventStream.NewConsumer(ctx, "worker", api.EventTypeJobCreated)
				if err != nil {
					return err
				}

				blobstore, err := bucket.New(ctx, bucketAddr)
				if err != nil {
					return err
				}

				wrkr, err := worker.New(ctx, workingDir, datastore, blobstore)
				if err != nil {
					return err
				}

				eventC, errC := eventStreamConsumer.Listen(ctx)

				logr.Info("listening for jobs")
				for {
					select {
					case err := <-errC:
						logr.Error(err, "event stream errored")
						return err
					case event := <-eventC:
						go func() {
							id := api.JobEventMetadata(event.Metadata).GetId()

							if err = wrkr.DoJob(ctx, id); err != nil {
								logr.Error(err, "job failed", "id", id)
								if err = eventStreamConsumer.Nack(event); err != nil {
									logr.Error(err, "failed to nack", "event", event.GetId())
								}
							} else {
								if err = eventStreamConsumer.Ack(event); err != nil {
									logr.Error(err, "failed to ack", "event", event.GetId())
								}
							}
						}()
					}
				}
			},
		}
	)

	workerCmd.Flags().StringVar(&amqpAddr, "amqp-addr", "", "AMQP address")
	workerCmd.Flags().StringVar(&bucketAddr, "bucket-addr", "", "bucket address")
	workerCmd.Flags().StringVar(&postgresAddr, "postgres-addr", "", "Postgres address")
	workerCmd.Flags().StringVar(&workingDir, "working-dir", "/var/lib/rototiller", "working directory")

	return workerCmd
}
