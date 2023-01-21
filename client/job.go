package client

import (
	"path"
	"time"

	"github.com/logsquaredn/rototiller/pb"
)

func (c *Client) GetJobs() ([]*pb.Job, error) {
	var (
		url  = c.url
		jobs = []*pb.Job{}
	)

	url.Path = pb.EndpointJobs

	return jobs, c.get(url, &jobs)
}

func (c *Client) GetJob(id string) (*pb.Job, error) {
	var (
		url = c.url
		job = &pb.Job{}
	)

	url.Path = path.Join(pb.EndpointJobs, id)

	return job, c.get(url, job)
}

func (c *Client) CreateJob(rawTaskType string, r Request) (*pb.Job, error) {
	var (
		url           = c.url
		job           = &pb.Job{}
		taskType, err = pb.ParseTaskType(rawTaskType)
	)
	if err != nil {
		return nil, err
	}

	url.Path = path.Join(pb.EndpointJobs, taskType.String())
	values := url.Query()
	for k, v := range r.Query() {
		if k != "" && v != "" {
			values.Add(k, v)
		}
	}
	url.RawQuery = values.Encode()

	return job, c.post(url, r, r.ContentType(), job)
}

func (c *Client) RunJob(rawTaskType string, r Request) (*pb.Job, error) {
	job, err := c.CreateJob(rawTaskType, r)
	if err != nil {
		return nil, err
	}

	for ; err == nil; job, err = c.GetJob(job.GetId()) {
		time.Sleep(c.pollInterval)
		switch job.Status {
		case pb.JobStatusComplete.String(), pb.JobStatusError.String():
			return job, nil
		}
	}

	return nil, err
}
