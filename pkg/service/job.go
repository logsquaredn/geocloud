package service

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/frantjc/go-js"
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller/pkg/api"
)

const (
	qInput    = "input"
	qInputOf  = "input-of"
	qOutputOf = "output-of"
)

func (a *API) createJobForCustomer(ctx *gin.Context, taskType api.TaskType, customer *api.Customer) (*api.Job, error) {
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
		storage *api.Storage
	)
	switch {
	case len(inputIDs) > 1:
		return nil, api.NewErr(fmt.Errorf("cannot specify more than one of queries '%s', '%s' and '%s'", qInput, qInputOf, qOutputOf), http.StatusBadRequest)
	case input != "":
		storage, err = a.getStorageForCustomer(input, customer)
		if err != nil {
			return nil, err
		}
	case inputOf != "":
		storage, err = a.getJobInputStorageForCustomer(ctx, inputOf, customer)
		if err != nil {
			return nil, err
		}
	case outputOf != "":
		storage, err = a.getJobOutputStorageForCustomer(ctx, outputOf, customer)
		if err != nil {
			return nil, err
		}
	default:
		storage, err = a.putRequestVolumeForCustomer(ctx, ctx.GetHeader("Content-Type"), ctx.Query("name"), ctx.Request.Body, customer)
		if err != nil {
			return nil, err
		}

		defer ctx.Request.Body.Close()
	}

	switch api.StorageStatus(storage.Status) {
	case api.StorageStatusFinal:
		return nil, api.NewErr(fmt.Errorf("cannot create job, storage id %s is final", storage.Id), http.StatusBadRequest)
	case api.StorageStatusUnusable:
		return nil, api.NewErr(fmt.Errorf("cannot create job, storage id %s is unsusable", storage.Id), http.StatusBadRequest)
	}

	job, err := a.Datastore.CreateJob(&api.Job{
		TaskType:   task.Type,
		Args:       buildJobArgs(ctx, task.Params),
		CustomerId: customer.Id,
		InputId:    storage.Id,
	})
	if err != nil {
		return nil, err
	}

	if err = a.EventStreamProducer.Emit(ctx, &api.Event{
		Type: api.EventTypeJobCreated.String(),
		Metadata: map[string]string{
			"id": job.Id,
		},
	}); err != nil {
		return nil, err
	}

	return job, nil
}

func (a *API) createJob(ctx *gin.Context, taskType api.TaskType) (*api.Job, error) {
	return a.createJobForCustomer(ctx, taskType, a.getAssumedCustomerFromContext(ctx))
}

func (a *API) getJob(ctx *gin.Context, id string) (*api.Job, error) {
	return a.getJobForCustomer(ctx, id, a.getAssumedCustomerFromContext(ctx))
}

func (a *API) getJobForCustomer(ctx *gin.Context, id string, customer *api.Customer) (*api.Job, error) {
	job, err := a.Datastore.GetJob(id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, api.NewErr(fmt.Errorf("job '%s' not found", id), http.StatusNotFound)
	case err != nil:
		return nil, err
	}

	return a.checkJobOwnershipForCustomer(job, customer)
}

func (a *API) checkJobOwnershipForCustomer(job *api.Job, customer *api.Customer) (*api.Job, error) {
	if job.CustomerId != customer.Id {
		return nil, api.NewErr(fmt.Errorf("customer does not own job '%s'", job.Id), http.StatusForbidden)
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
