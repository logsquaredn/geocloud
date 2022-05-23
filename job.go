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

func (c *Client) CreateJob(rawTaskType string, b []byte) (*Job, error) {
	var (
		url           = c.baseURL
		job           = &Job{}
		taskType, err = TaskTypeFrom(rawTaskType)
	)
	if err != nil {
		return nil, err
	}

	url.Path = path.Join(EndpointJob, taskType.String())

	return job, c.post(url, b, job)
}

func (c *Client) RunJob(rawTaskType string, b []byte) (*Job, error) {
	job, err := c.CreateJob(rawTaskType, b)
	if err != nil {
		return nil, err
	}

	for ; err == nil; job, err = c.GetJob(job.ID) {
		time.Sleep(time.Second)
		switch job.Status {
		case JobStatusComplete, JobStatusError:
			return job, nil
		}
	}

	return nil, err
}
