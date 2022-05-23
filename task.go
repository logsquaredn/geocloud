package geocloud

import (
	"fmt"
	"path"
	"strings"
)

type TaskType string

const (
	TaskTypeBuffer            TaskType = "buffer"
	TaskTypeFilter            TaskType = "filter"
	TaskTypeRemoveBadGeometry TaskType = "removebadgeometry"
	TaskTypeReproject         TaskType = "reproject"
	TaskTypeVectorLookup      TaskType = "vectorlookup"
	TaskTypeRasterLookup      TaskType = "rasterlookup"
)

var (
	AllTaskTypes = []TaskType{
		TaskTypeBuffer, TaskTypeFilter, TaskTypeRemoveBadGeometry,
		TaskTypeReproject, TaskTypeVectorLookup, TaskTypeRasterLookup,
	}
)

func (t TaskType) Type() string {
	return string(t)
}

func (t TaskType) Name() string {
	return t.Type()
}

func (t TaskType) String() string {
	return t.Type()
}

func ParseTaskType(taskType string) (TaskType, error) {
	for _, t := range AllTaskTypes {
		if strings.EqualFold(taskType, t.String()) {
			return t, nil
		}
	}
	return "", fmt.Errorf("unknown task type %s", taskType)
}

type Task struct {
	Type    TaskType `json:"type,omitempty"`
	Params  []string `json:"params,omitempty"`
	QueueID string   `json:"-"`
}

func (c *Client) GetTasks() ([]*Task, error) {
	var (
		url   = c.baseURL
		tasks = []*Task{}
	)

	url.Path = EndpointTask

	return tasks, c.get(url, &tasks)
}

func (c *Client) GetTask(rawTaskType string) (*Task, error) {
	var (
		url           = c.baseURL
		task          = &Task{}
		taskType, err = ParseTaskType(rawTaskType)
	)
	if err != nil {
		return nil, err
	}

	url.Path = path.Join(EndpointTask, taskType.String())

	return task, c.get(url, task)
}

func (c *Client) GetJobTask(id string) (*Task, error) {
	var (
		url  = c.baseURL
		task = &Task{}
	)

	url.Path = path.Join(EndpointJob, id, "task")

	return task, c.get(url, task)
}
