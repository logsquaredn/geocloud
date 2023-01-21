package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/logsquaredn/rototiller/pb"
)

func (a *Handler) getTask(rawTaskType string) (*pb.Task, error) {
	taskType, err := pb.ParseTaskType(rawTaskType)
	if err != nil {
		return nil, pb.NewErr(err, http.StatusBadRequest)
	}

	return a.getTaskType(taskType)
}

func (a *Handler) getTaskType(taskType pb.TaskType) (*pb.Task, error) {
	task, err := a.Datastore.GetTask(taskType)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, pb.NewErr(err, http.StatusNotFound)
	case err != nil:
		return nil, err
	}

	return task, nil
}
