package geocloud

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/jessevdk/go-flags"
	"github.com/tedsuo/ifrit"
)

// Component is a combination of a long-running process (e.g. containerd),
// a client to talk to said long-running process (e.g. a postgres client),
// and/or an API wrapped around said client to make interacting with it easy
// from other Components
//
// A Component must be started either by calling Component.Run or Component.Execute
// before its API should be expected to by functional
//
// geocloud, therefore, is an orchestration of Components that are ran in parallel
// and use one another's APIs to talk amongst themselves
// This makes each Component (datastore, objectstore, runtime, etc) pluggable
// (e.g. the messagequeue component could have an SQS and a RabbitMQ implementation)
type Component interface {
	flags.Commander
	ifrit.Runner

	// Name returns the name of the Component (e.g. "containerd")
	Name() string
	// IsConfigured returns whether or not the Component has all the configuration
	// necessary for it to expect to be able to run properly
	// (e.g. does the postgres datastore have a host to connect to)
	IsConfigured() bool
}

// AWSComponent is a Component that takes advantage of an
// *github.com/aws/aws-sdk-go/aws.Config
// so that such a Config can be shared by multiple AWSComponents
type AWSComponent interface {
	Component

	// WithConfig sets the AWSComponents' Config to the provided Config
	// and returns the AWSComponent for chaining
	WithConfig(*aws.Config) AWSComponent
}

// Message is the primary object that is passed between geocloud Components
type Message interface {
	ID() string
}

// JobStatus is an enum representing different statuses for geocloud Jobs
type JobStatus int64

const (
	Waiting JobStatus = iota
	InProgress
	Complete
	Error
)

// Status returns the string representation of a geocloud Job's status
func (s JobStatus) Status() string {
	switch s {
	case Waiting:
		return "waiting"
	case InProgress:
		return "inprogress"
	case Complete:
		return "complete"
	case Error:
		return "error"
	}
	return "unknown"
}

// JobStatusFrom creates a JobStatus from that JobStatus' string representation
func JobStatusFrom(jobStatus string) (JobStatus, error) {
	switch {
	case strings.EqualFold(Waiting.Status(), jobStatus):
		return Waiting, nil
	case strings.EqualFold(InProgress.Status(), jobStatus):
		return InProgress, nil
	case strings.EqualFold(Complete.Status(), jobStatus):
		return Complete, nil
	case strings.EqualFold(Error.Status(), jobStatus):
		return Error, nil
	}
	return -1, fmt.Errorf("unknown job status %s", jobStatus)
}

// String returns the string representation of a geocloud Job's status
func (s JobStatus) String() string {
	return s.Status()
}

// Job ...
type Job struct {
	Id        string
	TaskType  TaskType
	Status    JobStatus
	Err       error
	StartTime time.Time
	EndTime   time.Time
	Args      []string
}

var _ Message = (*Job)(nil)

// ID returns a Job's id
func (j *Job) ID() string {
	return j.Id
}

// File ...
type File interface {
	io.Reader

	// Name returns the path to the file relative to the
	// File's Volume
	Name() string
	Size() int
}

// WalkVolFunc ...
type WalkVolFunc func(string, File, error) error

// Volume ...
type Volume interface {
	// Walk iterates over each File in the Volume,
	// calling WalkVolFunc for each one
	Walk(WalkVolFunc) error
	// Download copies the contents of each File in the Volume
	// to the directory at the given path
	Download(string) error
}

// TaskType is a enum representing different types for geocloud Tasks
type TaskType int64

const (
	Buffer TaskType = iota
	Filter
	RemoveBadGeometry
	Reproject
)

// Type returns the string representation of a geocloud Task's type
func (t TaskType) Type() string {
	switch t {
	case Buffer:
		return "buffer"
	case Filter:
		return "filter"
	case RemoveBadGeometry:
		return "removebadgeometry"
	case Reproject:
		return "reproject"
	}
	return "unknown"
}

// Name returns the string representation of a geocloud Task's type
func (t TaskType) Name() string {
	return t.Type()
}

// String returns the string representation of a geocloud Task's type
func (t TaskType) String() string {
	return t.Type()
}

// TaskTypeFrom creates a TaskType from that TaskType's string representation
func TaskTypeFrom(taskType string) (TaskType, error) {
	switch {
	case strings.EqualFold(Buffer.Name(), taskType):
		return Buffer, nil
	case strings.EqualFold(Filter.Name(), taskType):
		return Filter, nil
	case strings.EqualFold(RemoveBadGeometry.Name(), taskType):
		return RemoveBadGeometry, nil
	case strings.EqualFold(Reproject.Name(), taskType):
		return Reproject, nil
	}
	return -1, fmt.Errorf("unknown task type %s", taskType)
}

// Task ...
type Task struct {
	Type    TaskType
	Params  []string
	QueueID string
	Ref     string
}

// Datastore ...
type Datastore interface {
	Component

	CreateJob(*Job) (*Job, error)
	UpdateJob(*Job) (*Job, error)
	GetTaskByJobID(Message) (*Task, error)
	GetTask(TaskType) (*Task, error)
	GetTasks(...TaskType) ([]*Task, error)
	GetJob(Message) (*Job, error)
	Migrate() error
}

// Objectstore ...
type Objectstore interface {
	Component

	GetInput(Message) (Volume, error)
	GetOutput(Message) (Volume, error)
	PutInput(Message, Volume) error
	PutOutput(Message, Volume) error
}

// MessageRecipient ...
type MessageRecipient interface {
	Component

	Send(Message) error
}

// MessageQueue ...
type MessageQueue interface {
	MessageRecipient

	WithDatastore(Datastore) MessageQueue
	WithMessageRecipient(Runtime) MessageQueue
	WithTasks(...TaskType) MessageQueue
	WithTaskmap(map[TaskType]string) MessageQueue
}

// API ...
type API interface {
	Component

	// Create(TaskType, []string) (Job, error)
	// Status(Message) (Job, error)
	// Result(Message) (map[string]string, error)

	WithDatastore(Datastore) API
	WithObjectstore(Objectstore) API
	WithMessageRecipient(MessageRecipient) API
}

// Runtime is a Component that ultimately recieves a
// Message, aggregates that Message's data from a
// Datastore and Objectstore, executes that
// Message's job, and sends the results back to the Datastore
// and Objectstore
type Runtime interface {
	// Runtime could substitue as MessageRecipient
	// since it ultimately is the recipient of messages;
	// MessageQueues such as SQS are just proxies
	//
	// With that said, for HA purposes, a genuine MessageQueue
	// should be utilized instead
	MessageRecipient

	WithDatastore(Datastore) Runtime
	WithObjectstore(Objectstore) Runtime
	WithWorkdir(string) Runtime
}
