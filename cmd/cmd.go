package main

import (
	"github.com/logsquaredn/geocloud/api/apicmd"
	"github.com/logsquaredn/geocloud/infrastructure/infrastructurecmd"
	"github.com/logsquaredn/geocloud/migrate/migratecmd"
	"github.com/logsquaredn/geocloud/worker/workercmd"
)

// GeocloudCmd is groups all of geocloud's subcommands and options under one binary
type GeocloudCmd struct {
	Version    func()     `long:"version" short:"v" description:"Print the version"`

	API    apicmd.APICmd       `command:"api" description:"Run the api"`  
	Worker workercmd.WorkerCmd `command:"worker" description:"Run the worker"`

	MigrateCmd      migratecmd.MigrateCmd               `command:"migrate" description:"Apply database migrations"`
	Infrastrcuture  infrastructurecmd.InfrastructureCmd `command:"infrastructure" alias:"infra" description:"Set up infrastructure"`

	// Quickstart      QuickstartCmd      `command:"quickstart" alias:"qs" description:"api, worker, and infrastructure commands if infrastructure configuration is not provided`
	// Keygen          KeygenCmd          `command:"keygen" alias:"kg" description:"generate RSA key for use with api or worker"`
}
