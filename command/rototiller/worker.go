package command

import (
	"os"
	"strconv"

	"github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/pb"
	"github.com/logsquaredn/rototiller/store/blob/bucket"
	"github.com/logsquaredn/rototiller/store/data/postgres"
	"github.com/logsquaredn/rototiller/stream/event/amqp"
	"github.com/logsquaredn/rototiller/worker"
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

				eventStreamConsumer, err := eventStream.NewConsumer(ctx, "worker", pb.EventTypeJobCreated)
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

				gorolimitVar := os.Getenv("GORO_LIMIT")
				gorolimit := 16
				if gorolimitVar != "" {
					gorolimit, err = strconv.Atoi(gorolimitVar)
					if err != nil {
						gorolimit = 16
					}
				}
				sem := make(chan struct{}, gorolimit)
				eventC, errC := eventStreamConsumer.Listen(ctx)

				logr.Info("listening for jobs")
				for {
					select {
					case err := <-errC:
						logr.Error(err, "event stream errored")
						return err
					case event := <-eventC:
						sem <- struct{}{}
						go func() {
							id := pb.JobEventMetadata(event.Metadata).GetId()

							if err = wrkr.DoJob(ctx, id); err != nil {
								logr.Error(err, "job failed", "id", id)
							}

							if err = eventStreamConsumer.Ack(event); err != nil {
								logr.Error(err, "failed to ack", "event", event.GetId())
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
