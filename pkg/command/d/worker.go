package command

import (
	"os"
	"strconv"

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

				eventStreamProducer, err := eventStream.NewProducer(ctx)
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

				var (
					eventC, errC = eventStreamConsumer.Listen(ctx)
					sem          = make(chan byte, 16)
				)
				if val := os.Getenv("GORO_LIMIT"); val != "" {
					if lim, err := strconv.Atoi(val); err == nil {
						sem = make(chan byte, lim)
					}
				}

				logr.Info("listening for jobs")
				for {
					select {
					case err := <-errC:
						logr.Error(err, "event stream errored")
						return err
					case event := <-eventC:
						sem <- 0

						go func() {
							id := api.JobEventMetadata(event.Metadata).GetId()

							if err = wrkr.DoJob(ctx, id); err != nil {
								logr.Error(err, "job failed", "id", id)
								if err = eventStreamConsumer.Nack(event); err != nil {
									logr.Error(err, "failed to nack event", "id", event.GetId())
								}
							} else {
								if err = eventStreamConsumer.Ack(event); err != nil {
									logr.Error(err, "failed to ack event", "id", event.GetId())
								}
							}

							if err = eventStreamProducer.Emit(ctx, &api.Event{
								Type:     api.EventTypeJobCompleted.String(),
								Metadata: event.Metadata,
							}); err != nil {
								logr.Error(err, "failed to emit event", "job", id, "type", api.EventTypeJobCompleted.String())
							}

							<-sem
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
