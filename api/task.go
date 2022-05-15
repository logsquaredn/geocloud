package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
)

func (a *API) getTask(ctx *gin.Context, rawTaskType string) (*geocloud.Task, int, error) {
	taskType, err := geocloud.TaskTypeFrom(rawTaskType)
	if err != nil {
		return nil, http.StatusNotFound, err
	}

	return a.getTaskType(ctx, taskType)
}

func (a *API) getTaskType(ctx *gin.Context, taskType geocloud.TaskType) (*geocloud.Task, int, error) {
	task, err := a.ds.GetTask(taskType)
	switch {
	case err == sql.ErrNoRows:
		return nil, http.StatusBadRequest, err
	case err != nil:
		return nil, http.StatusInternalServerError, err
	}

	return task, 0, nil
}
