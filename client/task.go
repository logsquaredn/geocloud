package client

import (
	"path"

	"github.com/logsquaredn/rototiller/pb"
)

func (c *Client) GetTasks() ([]*pb.Task, error) {
	var (
		url   = c.url
		tasks = []*pb.Task{}
	)

	url.Path = pb.EndpointTasks

	return tasks, c.get(url, &tasks)
}

func (c *Client) GetTask(rawTaskType string) (*pb.Task, error) {
	var (
		url           = c.url
		task          = &pb.Task{}
		taskType, err = pb.ParseTaskType(rawTaskType)
	)
	if err != nil {
		return nil, err
	}

	url.Path = path.Join(pb.EndpointTasks, taskType.String())

	return task, c.get(url, task)
}

func (c *Client) GetJobTask(id string) (*pb.Task, error) {
	var (
		url  = c.url
		task = &pb.Task{}
	)

	url.Path = path.Join(pb.EndpointJobs, id, "task")

	return task, c.get(url, task)
}
