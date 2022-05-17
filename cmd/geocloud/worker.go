package main

import (
	"github.com/logsquaredn/geocloud/datastore"
	"github.com/logsquaredn/geocloud/messagequeue"
	"github.com/logsquaredn/geocloud/objectstore"
	"github.com/logsquaredn/geocloud/worker"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use:     "worker",
	Aliases: []string{"w"},
	RunE:    runWorker,
}

var (
	workDir string
)

func init() {
	workerCmd.Flags().StringVar(&workDir, "workdir", "/var/lib/geocloud", "Working directory")
}

func runWorker(cmd *cobra.Command, args []string) error {
	ds, err := datastore.NewPostgres(
		getPostgresOpts(),
	)
	if err != nil {
		return err
	}

	os, err := objectstore.NewS3(
		getS3Opts(),
	)
	if err != nil {
		return err
	}

	rt, err := worker.New(
		&worker.Opts{
			Datastore:   ds,
			Objectstore: os,
			WorkDir:     workDir,
		},
	)
	if err != nil {
		return err
	}

	mq, err := messagequeue.NewAMQP(
		getAMQPOpts(),
	)
	if err != nil {
		return err
	}

	log.Info().Msg("polling for messages")
	return mq.Poll(rt.Send)
}
