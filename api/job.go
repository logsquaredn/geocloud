package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
)

func (a *API) createJobForCustomer(ctx *gin.Context, taskType geocloud.TaskType, customer *geocloud.Customer) (*geocloud.Job, int, error) {
	task, statusCode, err := a.getTaskType(ctx, taskType)
	if err != nil {
		return nil, statusCode, err
	}

	inputID := ctx.Query("input_id")
	if inputID == "" {
		storage, statusCode, err := a.putRequestVolumeForCustomer(ctx, customer)
		if err != nil {
			return nil, statusCode, err
		}
		inputID = storage.ID
	} else {
		_, statusCode, err := a.getStorageForCustomer(ctx, geocloud.NewMessage(inputID), customer)
		if err != nil {
			return nil, statusCode, err
		}
	}

	job, err := a.ds.CreateJob(&geocloud.Job{
		TaskType:   task.Type,
		Args:       buildJobArgs(ctx, task.Params),
		CustomerID: customer.ID,
		InputID:    inputID,
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
	case err == sql.ErrNoRows:
		ctx.AbortWithStatus(http.StatusNotFound)
	case err != nil:
		ctx.AbortWithStatus(http.StatusInternalServerError)
	}

	return a.checkJobOwnershipForCustomer(ctx, job, customer)
}

func (a *API) checkJobOwnershipForCustomer(ctx *gin.Context, job *geocloud.Job, customer *geocloud.Customer) (*geocloud.Job, int, error) {
	if job.CustomerID != customer.ID {
		// you could make an argument for 404 here
		// as "idk what storage you're talking about"
		// is more secure than "there's definitely a storage here
		// but you can't see it"
		return nil, http.StatusForbidden, fmt.Errorf("customer does not own job")
	}

	return job, 0, nil
}
