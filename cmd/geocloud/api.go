package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/logsquaredn/geocloud/api"
	"github.com/logsquaredn/geocloud/datastore"
	"github.com/logsquaredn/geocloud/messagequeue"
	"github.com/logsquaredn/geocloud/objectstore"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var apiCmd = &cobra.Command{
	Use:     "api",
	Aliases: []string{"a"},
	RunE:    runAPI,
}

var (
	port int64
)

func init() {
	apiCmd.Flags().Int64Var(&port, "port", 8080, "Listen port")
}

func runAPI(cmd *cobra.Command, args []string) error {
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

	mq, err := messagequeue.NewAMQP(
		getAMQPOpts(),
	)
	if err != nil {
		return err
	}

	srv, err := api.NewServer(&api.ServerOpts{
		Datastore:    ds,
		Objectstore:  os,
		MessageQueue: mq,
	})
	if err != nil {
		return err
	}

	addr := fmt.Sprintf(":%d", port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Info().Msgf("serving on %s", addr)
	return http.Serve(l, srv)
}
