package client

import (
	"path"
	"time"

	"github.com/logsquaredn/rototiller/pkg/api"
)

func (c *Client) GetJobs() ([]*api.Job, error) {
	var (
		url  = c.url
		jobs = []*api.Job{}
	)

	url.Path = api.EndpointJobs

	return jobs, c.get(url, &jobs)
}

func (c *Client) GetJob(id string) (*api.Job, error) {
	var (
		url = c.url
		job = &api.Job{}
	)

	url.Path = path.Join(api.EndpointJobs, id)

	return job, c.get(url, job)
}

func (c *Client) CreateJob(rawTaskType string, r Request) (*api.Job, error) {
	var (
		url           = c.url
		job           = &api.Job{}
		taskType, err = api.ParseTaskType(rawTaskType)
	)
	if err != nil {
		return nil, err
	}

	url.Path = path.Join(api.EndpointJobs, taskType.String())
	values := url.Query()
	for k, v := range r.Query() {
		if k != "" && v != "" {
			values.Add(k, v)
		}
	}
	url.RawQuery = values.Encode()

	return job, c.post(url, r, r.ContentType(), job)
}

func (c *Client) RunJob(rawTaskType string, r Request) (*api.Job, error) {
	job, err := c.CreateJob(rawTaskType, r)
	if err != nil {
		return nil, err
	}

	for ; err == nil; job, err = c.GetJob(job.GetId()) {
		time.Sleep(c.pollInterval)
		switch job.Status {
		case api.JobStatusComplete.String(), api.JobStatusError.String():
			return job, nil
		}
	}

	return nil, err
}
