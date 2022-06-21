package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/frantjc/go-js"
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
	errv1 "github.com/logsquaredn/geocloud/api/err/v1"
)

const (
	qInput    = "input"
	qInputOf  = "input-of"
	qOutputOf = "output-of"
)

func (a *API) createJobForCustomer(ctx *gin.Context, taskType geocloud.TaskType, customer *geocloud.Customer) (*geocloud.Job, error) {
	task, err := a.getTaskType(taskType)
	if err != nil {
		return nil, err
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
		return nil, errv1.New(fmt.Errorf("cannot specify more than one of queries '%s', '%s' and '%s'", qInput, qInputOf, qOutputOf), http.StatusBadRequest)
	case input != "":
		storage, err = a.getStorageForCustomer(geocloud.Msg(input), customer)
		if err != nil {
			return nil, err
		}
	case inputOf != "":
		storage, err = a.getJobInputStorageForCustomer(ctx, geocloud.Msg(inputOf), customer)
		if err != nil {
			return nil, err
		}
	case outputOf != "":
		storage, err = a.getJobOutputStorageForCustomer(ctx, geocloud.Msg(outputOf), customer)
		if err != nil {
			return nil, err
		}
	default:
		storage, err = a.putRequestVolumeForCustomer(ctx.GetHeader("Content-Type"), ctx.Query("name"), ctx.Request.Body, customer)
		if err != nil {
			return nil, err
		}

		defer ctx.Request.Body.Close()
	}

	if strings.EqualFold(storage.Status.String(), geocloud.StorageStatusFinal.String()) {
		return nil, errv1.New(fmt.Errorf("cannot create job, storage id %s is final", storage.ID), http.StatusBadRequest)
	} else if strings.EqualFold(storage.Status.String(), geocloud.StorageStatusUnusable.String()) {
		return nil, errv1.New(fmt.Errorf("cannot create job, storage id %s is unsusable", storage.ID), http.StatusBadRequest)
	}

	job, err := a.ds.CreateJob(&geocloud.Job{
		TaskType:   task.Type,
		Args:       buildJobArgs(ctx, task.Params),
		CustomerID: customer.ID,
		InputID:    storage.ID,
	})
	if err != nil {
		return nil, err
	}

	if err = a.mq.Send(job); err != nil {
		return nil, err
	}

	return job, nil
}

func (a *API) createJob(ctx *gin.Context, taskType geocloud.TaskType) (*geocloud.Job, error) {
	return a.createJobForCustomer(ctx, taskType, a.getAssumedCustomerFromContext(ctx))
}

func (a *API) getJob(ctx *gin.Context, m geocloud.Message) (*geocloud.Job, error) {
	return a.getJobForCustomer(ctx, m, a.getAssumedCustomerFromContext(ctx))
}

func (a *API) getJobForCustomer(ctx *gin.Context, m geocloud.Message, customer *geocloud.Customer) (*geocloud.Job, error) {
	job, err := a.ds.GetJob(m)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, errv1.New(fmt.Errorf("job '%s' not found", m.GetID()), http.StatusNotFound)
	case err != nil:
		return nil, err
	}

	return a.checkJobOwnershipForCustomer(job, customer)
}

func (a *API) checkJobOwnershipForCustomer(job *geocloud.Job, customer *geocloud.Customer) (*geocloud.Job, error) {
	if job.CustomerID != customer.ID {
		return nil, errv1.New(fmt.Errorf("customer does not own job '%s'", job.ID), http.StatusForbidden)
	}

	return job, nil
}

func buildJobArgs(ctx *gin.Context, taskParams []string) []string {
	jobArgs := make([]string, len(taskParams))
	for i, param := range taskParams {
		jobArgs[i] = ctx.Query(param)
	}
	return jobArgs
}
