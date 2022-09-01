package client

import (
	"path"

	"github.com/logsquaredn/rototiller/pkg/api"
)

func (c *Client) GetTasks() ([]*api.Task, error) {
	var (
		url   = c.url
		tasks = []*api.Task{}
	)

	url.Path = api.EndpointTasks

	return tasks, c.get(url, &tasks)
}

func (c *Client) GetTask(rawTaskType string) (*api.Task, error) {
	var (
		url           = c.url
		task          = &api.Task{}
		taskType, err = api.ParseTaskType(rawTaskType)
	)
	if err != nil {
		return nil, err
	}

	url.Path = path.Join(api.EndpointTasks, taskType.String())

	return task, c.get(url, task)
}

func (c *Client) GetJobTask(id string) (*api.Task, error) {
	var (
		url  = c.url
		task = &api.Task{}
	)

	url.Path = path.Join(api.EndpointJobs, id, "task")

	return task, c.get(url, task)
}
