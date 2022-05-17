package geocloud

import (
	"fmt"
	"strings"
	"time"
)

// JobStatus is an enum representing different statuses for geocloud Jobs
type JobStatus string

const (
	JobStatusWaiting    JobStatus = "waiting"
	JobStatusInProgress JobStatus = "inprogress"
	JobStatusComplete   JobStatus = "complete"
	JobStatusError      JobStatus = "error"
)

// Status returns the string representation of a geocloud Job's status
func (s JobStatus) Status() string {
	return string(s)
}

// JobStatusFrom creates a JobStatus from that JobStatus' string representation
func JobStatusFrom(jobStatus string) (JobStatus, error) {
	for _, j := range []JobStatus{
		JobStatusWaiting, JobStatusInProgress,
		JobStatusComplete, JobStatusError,
	} {
		if strings.EqualFold(jobStatus, j.String()) {
			return j, nil
		}
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
	TaskTypeBuffer            TaskType = "buffer"
	TaskTypeFilter            TaskType = "filter"
	TaskTypeRemoveBadGeometry TaskType = "removebadgeometry"
	TaskTypeReproject         TaskType = "reproject"
	TaskTypeVectorLookup      TaskType = "vectorlookup"
)

var (
	AllTaskTypes = []TaskType{
		TaskTypeBuffer, TaskTypeFilter, TaskTypeRemoveBadGeometry,
		TaskTypeReproject, TaskTypeVectorLookup,
	}
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
	for _, t := range AllTaskTypes {
		if strings.EqualFold(taskType, t.String()) {
			return t, nil
		}
	}
	return "", fmt.Errorf("unknown task type %s", taskType)
}

// Task ...
type Task struct {
	Type    TaskType `json:"type,omitempty"`
	Params  []string `json:"params,omitempty"`
	QueueID string   `json:"-"`
}

// Storage ...
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

type Error struct {
	Message string `json:"message,omitempty"`
}
