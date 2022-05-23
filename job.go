package geocloud

import (
	"fmt"
	"path"
	"strings"
	"time"
)

type JobStatus string

const (
	JobStatusWaiting    JobStatus = "waiting"
	JobStatusInProgress JobStatus = "inprogress"
	JobStatusComplete   JobStatus = "complete"
	JobStatusError      JobStatus = "error"
)

func (s JobStatus) Status() string {
	return string(s)
}

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

func (s JobStatus) String() string {
	return s.Status()
}

type Job struct {
	ID         string    `json:"id,omitempty"`
	CustomerID string    `json:"-"`
	InputID    string    `json:"input_id,omitempty"`
	OutputID   string    `json:"output_id,omitempty"`
	TaskType   TaskType  `json:"task_type,omitempty"`
	Status     JobStatus `json:"status,omitempty"`
	Error      string    `json:"error,omitempty"`
	StartTime  time.Time `json:"start_time,omitempty"`
	EndTime    time.Time `json:"end_time,omitempty"`
	Args       []string  `json:"args,omitempty"`
}

var _ Message = (*Job)(nil)

func (j *Job) GetID() string {
	return j.ID
}

func (c *Client) GetJobs() ([]*Job, error) {
	var (
		url  = c.baseURL
		jobs = []*Job{}
	)

	url.Path = EndpointJob

	return jobs, c.get(url, &jobs)
}

func (c *Client) GetJob(id string) (*Job, error) {
	var (
		url = c.baseURL
		job = &Job{}
	)

	url.Path = path.Join(EndpointJob, id)

	return job, c.get(url, job)
}
