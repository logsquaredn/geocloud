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

func ParseJobStatus(jobStatus string) (JobStatus, error) {
	for _, j := range []JobStatus{
		JobStatusWaiting, JobStatusInProgress,
		JobStatusComplete, JobStatusError,
	} {
		if strings.EqualFold(jobStatus, j.String()) {
			return j, nil
		}
	}
	return "", fmt.Errorf("unknown job status '%s'", jobStatus)
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

	url.Path = EndpointJobs

	return jobs, c.get(url, &jobs)
}

func (c *Client) GetJob(id string) (*Job, error) {
	var (
		url = c.baseURL
		job = &Job{}
	)

	url.Path = path.Join(EndpointJobs, id)

	return job, c.get(url, job)
}

func (c *Client) CreateJob(rawTaskType string, r Request) (*Job, error) {
	var (
		url           = c.baseURL
		job           = &Job{}
		taskType, err = ParseTaskType(rawTaskType)
	)
	if err != nil {
		return nil, err
	}

	url.Path = path.Join(EndpointJobs, taskType.String())
	values := url.Query()
	for k, v := range r.Query() {
		if k != "" && v != "" {
			values.Add(k, v)
		}
	}
	url.RawQuery = values.Encode()

	return job, c.post(url, r, r.ContentType(), job)
}

func (c *Client) RunJob(rawTaskType string, r Request) (*Job, error) {
	job, err := c.CreateJob(rawTaskType, r)
	if err != nil {
		return nil, err
	}

	for ; err == nil; job, err = c.GetJob(job.ID) {
		time.Sleep(c.pollInterval)
		switch job.Status {
		case JobStatusComplete, JobStatusError:
			return job, nil
		}
	}

	return nil, err
}
