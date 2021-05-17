package main

import (
	"github.com/logsquaredn/geocloud/api/apicmd"
	"github.com/logsquaredn/geocloud/tasks/mock/mockcmd"
	"github.com/logsquaredn/geocloud/worker/workercmd"
)

// GeocloudCmd is groups all of geocloud's subcommands and options go buiunder one binary
type GeocloudCmd struct {
	Version func() `short:"v" long:"version" description:"print the version"`

	API    apicmd.APICmd       `command:"api" description:"run the api"`  
	Worker workercmd.WorkerCmd `command:"worker" description:"run the worker"`
	Mock   mockcmd.MockCmd     `command:"mock" description:"run a mock task"`
	// Infrastrcuture  InfrastrcutureCmd  `command:"infrastructure" alias:"infra" description:"set up infrastructure`
	// Quickstart      QuickstartCmd      `command:"quickstart" alias:"qs" description:"api, worker, and infrastructure commands if infrastructure configuration is not provided`
	// Keygen          KeygenCmd          `command:"keygen" alias:"kg" description:"generate RSA key for use with api or worker"`
}
