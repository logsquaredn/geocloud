package cmd

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jessevdk/go-flags"
	"github.com/logsquaredn/geocloud"
	"github.com/logsquaredn/geocloud/api"
	"github.com/logsquaredn/geocloud/component"
	"github.com/logsquaredn/geocloud/datastore"
	"github.com/logsquaredn/geocloud/messagequeue"
	"github.com/logsquaredn/geocloud/objectstore"
	"github.com/logsquaredn/geocloud/runtime"
	"github.com/rs/zerolog"
	"github.com/tedsuo/ifrit"
)

type Geocloud struct {
	Version  func() `long:"version" short:"v" description:"Print the version"`
	Loglevel string `long:"log-level" short:"l" default:"info" choice:"info" choice:"debug" choice:"info" choice:"warn" choice:"error" choice:"fatal" choice:"panic" description:"Geocloud log level"`

	AWSGroup AWSGroup `group:"AWS" namespace:"aws"`

	API     APIComponent     `command:"api" description:"Run the api component"`
	Migrate MigrateComponent `command:"migrate" alias:"mgrt" description:"Apply database migrations"`
	Worker  WorkerComponent  `command:"worker" alias:"wrkr" description:"Run the worker component"`
	// Infrastructure InfrastructureComponent `command:"infrastructure" alias:"infra" description:"Apply infrastructure changes"`
	// Quickstart QuickstartComponent `command:"quickstart" alias:"qs" description:"Run all geocloud components"`
}

func (g *Geocloud) SetLogLevel() error {
	loglevel, err := zerolog.ParseLevel(g.Loglevel)
	if err == nil {
		zerolog.SetGlobalLevel(loglevel)
	}
	return err
}

var GeocloudCmd = &Geocloud{ Version: geocloud.V }

type AWSGroup struct {
	AccessKeyID     string         `long:"access-key-id" env:"GEOCLOUD_ACCESS_KEY_ID" description:"AWS access key ID"`
	SecretAccessKey string         `long:"secret-access-key" env:"GEOCLOUD_SECRET_ACCESS_KEY" description:"AWS secret access key"`
	Region          string         `long:"region" default:"us-east-1" description:"AWS region"`
	Profile         string         `long:"profile" default:"default" description:"AWS profile"`
	SharedCreds     flags.Filename `long:"shared-credentials-file" default:"~/.aws/credentials" description:"Path to AWS shared credentials file"`
}

func (a *AWSGroup) Session() (*session.Session, error) {
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.StaticProvider{
				Value: credentials.Value{
					AccessKeyID:     a.AccessKeyID,
					SecretAccessKey: a.SecretAccessKey,
				},
			},
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{
				Filename: string(a.SharedCreds),
				Profile:  a.Profile,
			},
		},
	)
	cfg := aws.NewConfig().WithRegion(a.Region).WithCredentials(creds)
	return session.NewSession(cfg)
}

type APIComponent struct {
	RestAPI *api.GinAPI

	PostgresDatastore *datastore.PostgresDatastore `group:"Postgres" namespace:"postgres"`

	// this need not be configured for the APIComponent
	SQSMessageQueue *messagequeue.SQSMessageQueue

	S3Objectstore *objectstore.S3Objectstore `group:"S3" namespace:"s3"`
}

var _ geocloud.Component = (*APIComponent)(nil)

func (a *APIComponent) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	err := GeocloudCmd.SetLogLevel()
	if err != nil {
		return fmt.Errorf("unable to set log level")
	}

	ds := a.PostgresDatastore
	if !ds.IsConfigured() {
		return fmt.Errorf("no datastore configured")
	}
	
	// TODO handle this potential error
	// We don't want to crash necessarily as we may utilize only
	// non-AWS components
	sess, _ := GeocloudCmd.AWSGroup.Session()
	a.SQSMessageQueue = &messagequeue.SQSMessageQueue{}
	a.SQSMessageQueue.WithSession(sess)
	mq := a.SQSMessageQueue.WithDatastore(ds)
	if !mq.IsConfigured() {
		return fmt.Errorf("no messagequeue configured")
	}

	a.S3Objectstore.WithSession(sess)
	os := a.S3Objectstore
	if !os.IsConfigured() {
		return fmt.Errorf("no objectstore configured")
	}

	a.RestAPI = &api.GinAPI{}
	api := a.RestAPI.WithDatastore(ds).WithMessageRecipient(mq).WithObjectstore(os)
	if !api.IsConfigured() {
		return fmt.Errorf("no api configured")
	}

	cs := component.Group(ds, mq, os, api)

	close(ready)
	return <-ifrit.Invoke(cs).Wait()
}

func (a *APIComponent) Execute(_ []string) error {
	return <-ifrit.Invoke(a).Wait()
}

func (a *APIComponent) Name() string {
	return "api"
}

func (a *APIComponent) IsConfigured() bool {
	// we don't expect to compose this component with other components
	// so this is a noop for now
	return true
}

type WorkerComponent struct {
	Tasks []string `long:"task" short:"t" description:"Task types that this worker should execute"`

	ContainerdRuntime *runtime.ContainerdRuntime `group:"Containerd" namespace:"containerd"`

	PostgresDatastore *datastore.PostgresDatastore `group:"Postgres" namespace:"postgres"`

	SQSMessageQueue *messagequeue.SQSMessageQueue `group:"SQS" namespace:"sqs"`

	S3Objectstore *objectstore.S3Objectstore `group:"S3" namespace:"s3"`
}

var _ geocloud.Component = (*WorkerComponent)(nil)

func (w *WorkerComponent) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	err := GeocloudCmd.SetLogLevel()
	if err != nil {
		return fmt.Errorf("unable to set log level")
	}

	ds := w.PostgresDatastore
	if !ds.IsConfigured() {
		return fmt.Errorf("no datastore configured")
	}

	// TODO handle this potential error
	// We don't want to crash necessarily as we may utilize only
	// non-AWS components
	sess, _ := GeocloudCmd.AWSGroup.Session()
	w.S3Objectstore.WithSession(sess)
	os := w.S3Objectstore
	if !os.IsConfigured() {
		return fmt.Errorf("no objectstore configured")
	}

	rt := w.ContainerdRuntime.WithDatastore(ds).WithObjectstore(os)
	if !rt.IsConfigured() {
		return fmt.Errorf("no runtime configured")
	}

	tasks := make([]geocloud.TaskType, len(w.Tasks))
	for i, task := range w.Tasks {
		tasks[i], err = geocloud.TaskTypeFrom(task)
		if err != nil {
			return err
		}
	}

	w.SQSMessageQueue.WithSession(sess)
	mq := w.SQSMessageQueue.WithMessageRecipient(rt).WithDatastore(ds).WithTasks(tasks...)
	if !mq.IsConfigured() {
		return fmt.Errorf("no message queue configured")
	}

	cs := component.Group(ds, os, rt, mq)

	close(ready)
	return <-ifrit.Invoke(cs).Wait()
}

func (w *WorkerComponent) Execute(_ []string) error {
	return <-ifrit.Invoke(w).Wait()
}

func (w *WorkerComponent) Name() string {
	return "worker"
}

func (w *WorkerComponent) IsConfigured() bool {
	// we don't expect to compose this component with other components
	// so this is a noop for now
	return true
}

type MigrateComponent struct {
	PostgresDatastore *datastore.PostgresDatastore `group:"Postgres" namespace:"postgres"`
}

func (m *MigrateComponent) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	ds := m.PostgresDatastore
	if ds == nil {
		return fmt.Errorf("no datastore configured")
	}
	defer close(ready)
	return ds.Migrate()
}

func (m *MigrateComponent) Execute(_ []string) error {
	return <-ifrit.Invoke(m).Wait()
}

func (m *MigrateComponent) Name() string {
	return "migrate"
}

func (m *MigrateComponent) IsConfigured() bool {
	// we don't expect to compose this component with other components
	// so this is a noop for now
	return true
}
