package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/logsquaredn/geocloud/api"
	"github.com/logsquaredn/geocloud/datastore"
	"github.com/logsquaredn/geocloud/messagequeue"
	"github.com/logsquaredn/geocloud/objectstore"
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

	if err = ds.Prepare(); err != nil {
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

	srv, err := api.NewServer(&api.APIOpts{
		Datastore:    ds,
		Objectstore:  os,
		MessageQueue: mq,
	})
	if err != nil {
		return err
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	return http.Serve(l, srv)
}
