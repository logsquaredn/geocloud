package geocloud

import (
	"fmt"
	"strings"
	"time"
)

// JobStatus is an enum representing different statuses for geocloud Jobs
type JobStatus string

const (
	Waiting    JobStatus = "waiting"
	InProgress JobStatus = "inprogress"
	Complete   JobStatus = "complete"
	Error      JobStatus = "error"
)

// Status returns the string representation of a geocloud Job's status
func (s JobStatus) Status() string {
	return string(s)
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
	return "", fmt.Errorf("unknown job status %s", jobStatus)
}

// String returns the string representation of a geocloud Job's status
func (s JobStatus) String() string {
	return s.Status()
}

// Job ...
type Job struct {
	ID         string    `json:"id,omitempty"`
	CustomerID string    `json:"-"`
	InputID    string    `json:"input_id,omitempty"`
	OutputID   string    `json:"output_id,omitempty"`
	TaskType   TaskType  `json:"task_type,omitempty"`
	Status     JobStatus `json:"status,omitempty"`
	Err        error     `json:"error,omitempty"`
	StartTime  time.Time `json:"start_time,omitempty"`
	EndTime    time.Time `json:"end_time,omitempty"`
	Args       []string  `json:"args,omitempty"`
}

var _ Message = (*Job)(nil)

// GetID returns a Job's id
func (j *Job) GetID() string {
	return j.ID
}

type Customer struct {
	ID string `json:"id,omitempty"`
}

var _ Message = (*Customer)(nil)

// GetID returns a Customer's id
func (c *Customer) GetID() string {
	return c.ID
}

// TaskType is a enum representing different types for geocloud Tasks
type TaskType string

const (
	Buffer            TaskType = "buffer"
	Filter            TaskType = "filter"
	RemoveBadGeometry TaskType = "removebadgeometry"
	Reproject         TaskType = "reproject"
	VectorLookup      TaskType = "vectorlookup"
)

// Type returns the string representation of a geocloud Task's type
func (t TaskType) Type() string {
	return string(t)
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
	return "", fmt.Errorf("unknown task type %s", taskType)
}

// Task ...
type Task struct {
	Type    TaskType `json:"type,omitempty"`
	Params  []string `json:"params,omitempty"`
	QueueID string   `json:"-"`
}

// Job ...
type Storage struct {
	ID         string    `json:"id,omitempty"`
	CustomerID string    `json:"-"`
	Name       string    `json:"name,omitempty"`
	LastUsed   time.Time `json:"last_used,omitempty"`
}

var _ Message = (*Storage)(nil)

// GetID returns a Storage's id
func (s *Storage) GetID() string {
	return s.ID
}
