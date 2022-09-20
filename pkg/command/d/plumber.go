package command

import (
	"errors"

	"github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/pkg/api"
	"github.com/logsquaredn/rototiller/pkg/sink"
	"github.com/logsquaredn/rototiller/pkg/store/blob/bucket"
	"github.com/logsquaredn/rototiller/pkg/store/data/postgres"
	"github.com/logsquaredn/rototiller/pkg/stream/event/amqp"
	"github.com/logsquaredn/rototiller/pkg/volume"
	"github.com/spf13/cobra"
)

var (
	errDrained = errors.New("drained")
)

func NewPlumber() *cobra.Command {
	var (
		postgresAddr, bucketAddr, amqpAddr string
		plumberCmd                         = &cobra.Command{
			Use:     "plumber",
			Aliases: []string{"p"},
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

				eventStreamConsumer, err := eventStream.NewConsumer(ctx, "plumber", api.EventTypeJobCompleted)
				if err != nil {
					return err
				}

				blobstore, err := bucket.New(ctx, bucketAddr)
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

							sinks, err := datastore.GetJobSinks(id)
							if err != nil {
								logr.Error(err, "draining to sink failed", "id", id)
								if err = eventStreamConsumer.Nack(event); err != nil {
									logr.Error(err, "failed to nack", "event", event.GetId())
								}
							} else {
								for _, s := range sinks {
									storage, err := datastore.GetJobOutputStorage(id)
									if err != nil {
										logr.Error(err, "failed to get output", "job", id)
										break
									}

									vol, err := blobstore.GetObject(ctx, storage.GetId())
									if err != nil {
										logr.Error(err, "failed to read output", "job", id, "storage", storage.GetId())
										break
									}

									d, err := sink.OpenSink(ctx, s.GetAddress())
									if err != nil {
										logr.Error(err, "failed to open sink", "address", s.GetAddress())
										break
									}

									if err = vol.Walk(func(_ string, f volume.File, e error) error {
										// pass errors through
										// so we only drain once
										if e != nil {
											return e
										}

										if e = d.Drain(ctx, f); e == nil {
											return errDrained
										}

										return nil
									}); err != nil && !errors.Is(err, errDrained) {
										logr.Error(err, "failed to drain to sink", "address", s.GetAddress())
										break
									}

									if err = eventStreamConsumer.Ack(event); err != nil {
										logr.Error(err, "failed to ack", "event", event.GetId())
									}
								}
							}
						}()
					}
				}
			},
		}
	)

	plumberCmd.Flags().StringVar(&amqpAddr, "amqp-addr", "", "AMQP address")
	plumberCmd.Flags().StringVar(&bucketAddr, "bucket-addr", "", "bucket address")
	plumberCmd.Flags().StringVar(&postgresAddr, "postgres-addr", "", "Postgres address")

	return plumberCmd
}
