package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/logsquaredn/geocloud"
	errv1 "github.com/logsquaredn/geocloud/api/err/v1"
)

func (a *API) getTask(rawTaskType string) (*geocloud.Task, error) {
	taskType, err := geocloud.ParseTaskType(rawTaskType)
	if err != nil {
		return nil, errv1.New(err, http.StatusBadRequest)
	}

	return a.getTaskType(taskType)
}

func (a *API) getTaskType(taskType geocloud.TaskType) (*geocloud.Task, error) {
	task, err := a.ds.GetTask(taskType)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, errv1.New(err, http.StatusNotFound)
	case err != nil:
		return nil, err
	}

	return task, nil
}
