package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/frantjc/go-js"
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller/pb"
)

const (
	qInput    = "input"
	qInputOf  = "input-of"
	qOutputOf = "output-of"
)

func (a *Handler) createJobForNamespace(ctx *gin.Context, taskType pb.TaskType, namespace string) (*pb.Job, error) {
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
		storage *pb.Storage
	)
	switch {
	case len(inputIDs) > 1:
		return nil, pb.NewErr(fmt.Errorf("cannot specify more than one of queries '%s', '%s' and '%s'", qInput, qInputOf, qOutputOf), http.StatusBadRequest)
	case input != "":
		storage, err = a.getStorageForNamespace(input, namespace)
		if err != nil {
			return nil, err
		}
	case inputOf != "":
		storage, err = a.getJobInputStorageForNamespace(ctx, inputOf, namespace)
		if err != nil {
			return nil, err
		}
	case outputOf != "":
		storage, err = a.getJobOutputStorageForNamespace(ctx, outputOf, namespace)
		if err != nil {
			return nil, err
		}
	default:
		storage, err = a.putRequestVolumeForNamespace(ctx, ctx.GetHeader("Content-Type"), ctx.Query("name"), ctx.Request.Body, namespace)
		if err != nil {
			return nil, err
		}

		defer ctx.Request.Body.Close()
	}

	switch pb.StorageStatus(storage.Status) {
	case pb.StorageStatusFinal:
		return nil, pb.NewErr(fmt.Errorf("cannot create job, storage id %s is final", storage.Id), http.StatusBadRequest)
	case pb.StorageStatusUnusable:
		return nil, pb.NewErr(fmt.Errorf("cannot create job, storage id %s is unsusable", storage.Id), http.StatusBadRequest)
	}

	job, err := a.Datastore.CreateJob(&pb.Job{
		Steps: []*pb.Step{
			{
				TaskType: task.Type,
				Args:     buildJobArgs(ctx, task.Params),
			},
		},
		Namespace: namespace,
		InputId:   storage.Id,
	})
	if err != nil {
		return nil, err
	}

	if err = a.EventStreamProducer.Emit(ctx, &pb.Event{
		Type: pb.EventTypeJobCreated.String(),
		Metadata: map[string]string{
			"id": job.Id,
		},
	}); err != nil {
		return nil, err
	}

	return job, nil
}

func (a *Handler) createJob(ctx *gin.Context, taskType pb.TaskType) (*pb.Job, error) {
	namespace, err := a.getNamespaceFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return a.createJobForNamespace(ctx, taskType, namespace)
}

func (a *Handler) getJob(ctx *gin.Context, id string) (*pb.Job, error) {
	namespace, err := a.getNamespaceFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return a.getJobForNamespace(ctx, id, namespace)
}

func (a *Handler) getJobForNamespace(ctx *gin.Context, id string, namespace string) (*pb.Job, error) {
	job, err := a.Datastore.GetJob(id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, pb.NewErr(fmt.Errorf("job '%s' not found", id), http.StatusNotFound)
	case err != nil:
		return nil, err
	}

	return a.checkJobOwnership(job, namespace)
}

func (a *Handler) checkJobOwnership(job *pb.Job, namespace string) (*pb.Job, error) {
	if job.Namespace != namespace {
		return nil, pb.NewErr(fmt.Errorf("user does not own job '%s'", job.Id), http.StatusForbidden)
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
