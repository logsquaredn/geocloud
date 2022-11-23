package service

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/logsquaredn/rototiller/pkg/api"
)

func (a *Handler) getTask(rawTaskType string) (*api.Task, error) {
	taskType, err := api.ParseTaskType(rawTaskType)
	if err != nil {
		return nil, api.NewErr(err, http.StatusBadRequest)
	}

	return a.getTaskType(taskType)
}

func (a *Handler) getTaskType(taskType api.TaskType) (*api.Task, error) {
	task, err := a.Datastore.GetTask(taskType)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, api.NewErr(err, http.StatusNotFound)
	case err != nil:
		return nil, err
	}

	return task, nil
}

func (a *Handler) getTasksFromJobSteps(job *api.Job) ([]*api.Task, error) {
	var tasks []*api.Task
	for _, step := range job.Steps {
		task, err := a.getTaskType(api.TaskType(step.TaskType))
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}
