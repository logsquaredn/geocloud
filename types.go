package geocloud

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// Message is the primary object that is passed between geocloud components
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
	Id         string
	TaskType   TaskType
	Status     JobStatus
	Err        error
	StartTime  time.Time
	EndTime    time.Time
	Args       []string
	CustomerID string
}

var _ Message = (*Job)(nil)

// ID returns a Job's id
func (j *Job) ID() string {
	return j.Id
}

type Customer struct {
	Id   string
	Name string
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
}

// Datastore ...
type Datastore interface {
	CreateJob(*Job) (*Job, error)
	UpdateJob(*Job) (*Job, error)
	GetTaskByJobID(Message) (*Task, error)
	GetTask(TaskType) (*Task, error)
	GetTasks(...TaskType) ([]*Task, error)
	GetCustomer(string) (*Customer, error)
	GetJob(Message) (*Job, error)
	GetJobs(time.Duration) ([]*Job, error)
	DeleteJob(*Job) error
}

// Objectstore ...
type Objectstore interface {
	GetInput(Message) (Volume, error)
	GetOutput(Message) (Volume, error)
	PutInput(Message, Volume) error
	PutOutput(Message, Volume) error
	DeleteRecursive(string) error
}

// MessageRecipient ...
type MessageRecipient interface {
	Send(Message) error
}

// MessageQueue ...
type MessageQueue interface {
	MessageRecipient
	Poll(f func(Message) error) error
}

// API ...
type API interface {
	Serve(l net.Listener) error
}

// Runtime is the component that ultimately recieves a
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
}

type CreateResponse struct {
	Id string `json:"id"`
}

type StatusResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
