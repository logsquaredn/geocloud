package service

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/logsquaredn/rototiller/pkg/api"
)

func (a *API) getTask(rawTaskType string) (*api.Task, error) {
	taskType, err := api.ParseTaskType(rawTaskType)
	if err != nil {
		return nil, api.NewErr(err, http.StatusBadRequest)
	}

	return a.getTaskType(taskType)
}

func (a *API) getTaskType(taskType api.TaskType) (*api.Task, error) {
	task, err := a.Datastore.GetTask(taskType)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, api.NewErr(err, http.StatusNotFound)
	case err != nil:
		return nil, err
	}

	return task, nil
}
