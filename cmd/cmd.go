package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
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
	Loglevel string `long:"log-level" short:"l" default:"info" choice:"trace" choice:"debug" choice:"info" choice:"warn" choice:"error" choice:"fatal" choice:"panic" description:"Geocloud log level"`

	AWSGroup AWSGroup `group:"AWS" namespace:"aws"`

	API       APIComponent       `command:"api" alias:"a" description:"Run the api component"`
	Migrate   MigrateComponent   `command:"migrate" alias:"m" description:"Apply database migrations"`
	Worker    WorkerComponent    `command:"worker" alias:"w" description:"Run the worker component"`
	Secretary SecretaryComponent `command:"secretary" alias:"s" description:"Run the secretary component"`
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

var GeocloudCmd = &Geocloud{Version: geocloud.V}

type AWSGroup struct {
	AccessKeyID     string         `long:"access-key-id" env:"GEOCLOUD_ACCESS_KEY_ID" description:"AWS access key ID"`
	SecretAccessKey string         `long:"secret-access-key" env:"GEOCLOUD_SECRET_ACCESS_KEY" description:"AWS secret access key"`
	Region          string         `long:"region" default:"us-east-1" description:"AWS region"`
	Profile         string         `long:"profile" default:"default" description:"AWS profile"`
	SharedCreds     flags.Filename `long:"shared-credentials-file" default:"~/.aws/credentials" description:"Path to AWS shared credentials file"`
}

var home string

func init() {
	home = os.Getenv("HOME")
}

func (a *AWSGroup) Config() (*aws.Config, error) {
	accessKeyID, secretAccessKey := a.AccessKeyID, a.SecretAccessKey
	if accessKeyID == "" {
		accessKeyID = os.Getenv("GEOCLOUD_AWS_ACCESS_KEY_ID")
	}

	if secretAccessKey == "" {
		secretAccessKey = os.Getenv("GEOCLOUD_AWS_SECRET_ACCESS_KEY")
	}

	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.StaticProvider{
				Value: credentials.Value{
					AccessKeyID:     accessKeyID,
					SecretAccessKey: secretAccessKey,
				},
			},
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{
				Filename: strings.Replace(string(a.SharedCreds), "~", home, -1),
				Profile:  a.Profile,
			},
		},
	)
	return aws.NewConfig().WithRegion(a.Region).WithCredentials(creds), nil
}

type RegistryGroup struct {
	Username string `long:"username" env:"GEOCLOUD_REGISTRY_USERNAME" description:"Registry username"`
	Password string `long:"password" env:"GEOCLOUD_REGISTRY_PASSWORD" description:"Regsitry password"`
}

func (r *RegistryGroup) Resolver() (*remotes.Resolver, error) {
	authorizer := docker.NewDockerAuthorizer(
		docker.WithAuthCreds(func(string) (string, string, error) {
			return r.Username, r.Password, nil
		}),
	)
	opts := []docker.RegistryOpt{
		docker.WithAuthorizer(authorizer),
		docker.WithPlainHTTP(docker.MatchLocalhost),
	}
	resolver := docker.NewResolver(docker.ResolverOptions{
		Hosts: docker.ConfigureDefaultRegistries(opts...),
	})
	return &resolver, nil
}

type APIComponent struct {
	Tasks map[string]string `long:"task" short:"t" description:"Map of task types to queue IDs"`

	RestAPI *api.GinAPI `group:"Gin" namespace:"gin"`

	PostgresDatastore *datastore.PostgresDatastore `group:"Postgres" namespace:"postgres"`

	// these need not be configured for the APIComponent
	SQSMessageQueue  *messagequeue.SQSMessageQueue  `group:"SQS" namespace:"sqs"`
	AMQPMessageQueue *messagequeue.AMQPMessageQueue `group:"AMQP" namespace:"amqp"`

	S3Objectstore *objectstore.S3Objectstore `group:"S3" namespace:"s3"`
}

var _ geocloud.Component = (*APIComponent)(nil)

func (a *APIComponent) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	err := GeocloudCmd.SetLogLevel()
	if err != nil {
		return fmt.Errorf("unable to set log level")
	}

	ds := a.PostgresDatastore
	if !ds.IsEnabled() {
		return fmt.Errorf("no datastore enabled")
	}

	tm := map[geocloud.TaskType]string{}
	for taskType, id := range a.Tasks {
		task, err := geocloud.TaskTypeFrom(taskType)
		if err != nil {
			return err
		}
		tm[task] = id
	}
	cfg, _ := GeocloudCmd.AWSGroup.Config()
	mq, ok := component.Coalesce(a.SQSMessageQueue.WithConfig(cfg), a.AMQPMessageQueue).(geocloud.MessageQueue)
	if !ok || mq == nil {
		return fmt.Errorf("no messagequeue enabled")
	}

	os, ok := a.S3Objectstore.WithConfig(cfg).(geocloud.Objectstore)
	if !ok || !os.IsEnabled() {
		return fmt.Errorf("no objectstore enabled")
	}

	api := a.RestAPI.WithDatastore(ds).WithMessageRecipient(mq).WithObjectstore(os)
	if !api.IsEnabled() {
		return fmt.Errorf("no api enabled")
	}

	cs := component.NewGroup(ds, mq.WithDatastore(ds).WithTaskmap(tm), os, api)

	return cs.Run(signals, ready)
}

func (a *APIComponent) Execute(_ []string) error {
	return <-ifrit.Invoke(a).Wait()
}

func (a *APIComponent) Name() string {
	return "api"
}

func (a *APIComponent) IsEnabled() bool {
	return true
}

type WorkerComponent struct {
	Tasks   []string       `long:"task" short:"t" description:"Task types that this worker should execute"`
	Workdir flags.Filename `long:"workdir" short:"w" default:"/var/lib/geocloud" descirption:"Directory for geocloud to store working data in"`

	RegistryGroup *RegistryGroup `group:"Registry" namespace:"registry"`

	ContainerdRuntime *runtime.ContainerdRuntime `group:"Containerd" namespace:"containerd"`

	PostgresDatastore *datastore.PostgresDatastore `group:"Postgres" namespace:"postgres"`

	SQSMessageQueue  *messagequeue.SQSMessageQueue  `group:"SQS" namespace:"sqs"`
	AMQPMessageQueue *messagequeue.AMQPMessageQueue `group:"AMQP" namespace:"amqp"`

	S3Objectstore *objectstore.S3Objectstore `group:"S3" namespace:"s3"`
}

var _ geocloud.Component = (*WorkerComponent)(nil)

func (w *WorkerComponent) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	workdir := string(w.Workdir)

	err := GeocloudCmd.SetLogLevel()
	if err != nil {
		return fmt.Errorf("unable to set log level")
	}

	ds := w.PostgresDatastore
	if !ds.IsEnabled() {
		return fmt.Errorf("no datastore enabled")
	}

	cfg, _ := GeocloudCmd.AWSGroup.Config()
	os, ok := w.S3Objectstore.WithConfig(cfg).(geocloud.Objectstore)
	if !ok || !os.IsEnabled() {
		return fmt.Errorf("no objectstore enabled")
	}

	resolver, _ := w.RegistryGroup.Resolver()
	rt := w.ContainerdRuntime.WithResolver(resolver).WithWorkdir(workdir).WithDatastore(ds).WithObjectstore(os)
	if !rt.IsEnabled() {
		return fmt.Errorf("no runtime enabled")
	}

	tasks := make([]geocloud.TaskType, len(w.Tasks))
	for i, task := range w.Tasks {
		tasks[i], err = geocloud.TaskTypeFrom(task)
		if err != nil {
			return err
		}
	}

	mq, ok := component.Coalesce(w.SQSMessageQueue.WithConfig(cfg), w.AMQPMessageQueue).(geocloud.MessageQueue)
	if !ok || mq == nil {
		return fmt.Errorf("no messagequeue enabled")
	}

	cs := component.NewGroup(ds, os, rt, mq.WithDatastore(ds).WithMessageRecipient(rt).WithTasks(tasks...))

	return cs.Run(signals, ready)
}

func (w *WorkerComponent) Execute(_ []string) error {
	return <-ifrit.Invoke(w).Wait()
}

func (w *WorkerComponent) Name() string {
	return "worker"
}

func (w *WorkerComponent) IsEnabled() bool {
	return true
}

type MigrateComponent struct {
	PostgresDatastore *datastore.PostgresDatastore `group:"Postgres" namespace:"postgres"`
}

func (m *MigrateComponent) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	ds := m.PostgresDatastore
	if !ds.IsEnabled() {
		return fmt.Errorf("no datastore enabled")
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

func (m *MigrateComponent) IsEnabled() bool {
	return true
}

type SecretaryComponent struct {
	PostgresDatastore    *datastore.PostgresDatastore `group:"Postgres" namespace:"postgres"`
	S3Objectstore        *objectstore.S3Objectstore   `group:"S3" namespace:"s3"`
	S3ArchiveObjectstore *objectstore.S3Objectstore   `group:"S3Archive" namespace:"s3-archive"`
	WorkJobsBefore       time.Duration                `long:"work-jobs-before" default:"24h" description:"Work on jobs before this time"`
}

func (s *SecretaryComponent) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	dataStore := s.PostgresDatastore
	if !dataStore.IsEnabled() {
		return fmt.Errorf("no datastore enabled")
	}

	cfg, _ := GeocloudCmd.AWSGroup.Config()
	objStore, ok := s.S3Objectstore.WithConfig(cfg).(geocloud.Objectstore)
	if !ok || !objStore.IsEnabled() {
		return fmt.Errorf("no objectstore enabled")
	}

	cs := component.NewGroup(dataStore, objStore, component.NewComponentFunc(
		func(_ <-chan os.Signal, ready chan<- struct{}) error {
			jobs, err := dataStore.GetJobs(s.WorkJobsBefore)
			if err != nil {
				return err
			}

			for _, j := range jobs {
				fmt.Println(j.Id)
				fmt.Println(j.CustomerID)
			}

			close(ready)
			return nil
		},
	))

	return cs.Run(signals, ready)
}

func (s *SecretaryComponent) Execute(_ []string) error {
	return <-ifrit.Invoke(s).Wait()
}

func (s *SecretaryComponent) Name() string {
	return "secretary"
}

func (s *SecretaryComponent) IsEnabled() bool {
	return true
}
