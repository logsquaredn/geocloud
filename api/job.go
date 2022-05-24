package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/frantjc/go-js"
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
)

const (
	qInput    = "input"
	qInputOf  = "input-of"
	qOutputOf = "output-of"
)

func (a *API) createJobForCustomer(ctx *gin.Context, taskType geocloud.TaskType, customer *geocloud.Customer) (*geocloud.Job, int, error) {
	task, statusCode, err := a.getTaskType(ctx, taskType)
	if err != nil {
		return nil, statusCode, err
	}

	var (
		input    = ctx.Query(qInput)
		inputOf  = ctx.Query(qInputOf)
		outputOf = ctx.Query(qOutputOf)
		inputIDs = js.Filter(
			[]string{input, inputOf, outputOf},
			func(s string, _ int, _ []string) bool {
				return s != ""
			},
		)
		storage *geocloud.Storage
	)
	switch {
	case len(inputIDs) > 1:
		return nil, 400, fmt.Errorf("cannot specify more than one of queries '%s', '%s' and '%s'", qInput, qInputOf, qOutputOf)
	case input != "":
		storage, statusCode, err = a.getStorageForCustomer(ctx, geocloud.NewMessage(input), customer)
		if err != nil {
			return nil, statusCode, err
		}
	case inputOf != "":
		storage, statusCode, err = a.getJobInputStorageForCustomer(ctx, geocloud.NewMessage(inputOf), customer)
		if err != nil {
			return nil, statusCode, err
		}
	case outputOf != "":
		storage, statusCode, err = a.getJobOutputStorageForCustomer(ctx, geocloud.NewMessage(outputOf), customer)
		if err != nil {
			return nil, statusCode, err
		}
	default:
		storage, statusCode, err = a.putRequestVolumeForCustomer(ctx, customer)
		if err != nil {
			return nil, statusCode, err
		}
	}

	job, err := a.ds.CreateJob(&geocloud.Job{
		TaskType:   task.Type,
		Args:       buildJobArgs(ctx, task.Params),
		CustomerID: customer.ID,
		InputID:    storage.ID,
	})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	if err = a.mq.Send(job); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return job, 0, nil
}

func (a *API) createJob(ctx *gin.Context, taskType geocloud.TaskType) (*geocloud.Job, int, error) {
	return a.createJobForCustomer(ctx, taskType, a.getAssumedCustomer(ctx))
}

func (a *API) getJob(ctx *gin.Context, m geocloud.Message) (*geocloud.Job, int, error) {
	return a.getJobForCustomer(ctx, m, a.getAssumedCustomer(ctx))
}

func (a *API) getJobForCustomer(ctx *gin.Context, m geocloud.Message, customer *geocloud.Customer) (*geocloud.Job, int, error) {
	job, err := a.ds.GetJob(m)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, http.StatusNotFound, fmt.Errorf("job '%s' not found", m.GetID())
	case err != nil:
		return nil, http.StatusInternalServerError, err
	}

	return a.checkJobOwnershipForCustomer(job, customer)
}

func (a *API) checkJobOwnershipForCustomer(job *geocloud.Job, customer *geocloud.Customer) (*geocloud.Job, int, error) {
	if job.CustomerID != customer.ID {
		return nil, http.StatusForbidden, fmt.Errorf("customer does not own job '%s'", job.ID)
	}

	return job, 0, nil
}

func buildJobArgs(ctx *gin.Context, taskParams []string) []string {
	jobArgs := make([]string, len(taskParams))
	for i, param := range taskParams {
		jobArgs[i] = ctx.Query(param)
	}
	return jobArgs
}
