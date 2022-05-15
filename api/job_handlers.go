package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
	"github.com/rs/zerolog/log"
)

func (a *API) listJobHandler(ctx *gin.Context) {
	jobs, err := a.ds.GetCustomerJobs(a.getAssumedCustomer(ctx))
	switch {
	case err == sql.ErrNoRows:
		ctx.AbortWithStatus(http.StatusNotFound)
	case err != nil:
		ctx.AbortWithStatus(http.StatusInternalServerError)
	}

	ctx.JSON(http.StatusOK, jobs)
}

func (a *API) getJobHandler(ctx *gin.Context) {
	job, statusCode, err := a.getJob(ctx, geocloud.NewMessage(ctx.Param("id")))
	if err != nil {
		log.Err(err).Msg("unable to job")
		ctx.AbortWithStatus(statusCode)
		return
	}

	ctx.JSON(http.StatusOK, job)
}

func (a *API) getJobTaskHandler(ctx *gin.Context) {
	job, statusCode, err := a.getJob(ctx, geocloud.NewMessage(ctx.Param("id")))
	if err != nil {
		log.Err(err).Msg("unable to job")
		ctx.AbortWithStatus(statusCode)
		return
	}

	task, statusCode, err := a.getTaskType(ctx, job.TaskType)
	if err != nil {
		log.Err(err).Msg("unable to task")
		ctx.AbortWithStatus(statusCode)
		return
	}

	ctx.JSON(http.StatusOK, task)
}

func (a *API) getJobInputHandler(ctx *gin.Context) {
	storage, statusCode, err := a.getJobInputStorage(ctx, geocloud.NewMessage(ctx.Param("id")))
	if err != nil {
		log.Err(err).Msg("unable to job input storage")
		ctx.AbortWithStatus(statusCode)
		return
	}

	ctx.JSON(http.StatusOK, storage)
}

func (a *API) getJobInputContentHandler(ctx *gin.Context) {
	storage, statusCode, err := a.getJobInputStorage(ctx, geocloud.NewMessage(ctx.Param("id")))
	if err != nil {
		log.Err(err).Msg("unable to get storage")
		ctx.AbortWithStatus(statusCode)
		return
	}

	volume, err := a.os.GetObject(storage)
	if err != nil {
		log.Err(err).Msg("unable to get object")
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	b, contentType, statusCode, err := a.getVolumeContent(ctx, volume)
	if err != nil {
		log.Err(err).Msg("unable to get volume content")
		ctx.AbortWithStatus(statusCode)
		return
	}

	ctx.Data(http.StatusOK, contentType, b)
}

func (a *API) getJobOutputHandler(ctx *gin.Context) {
	storage, statusCode, err := a.getJobOutputStorage(ctx, geocloud.NewMessage(ctx.Param("id")))
	if err != nil {
		log.Err(err).Msg("unable to job output storage")
		ctx.AbortWithStatus(statusCode)
		return
	}

	ctx.JSON(http.StatusOK, storage)
}

func (a *API) getJobOutputContentHandler(ctx *gin.Context) {
	storage, statusCode, err := a.getJobOutputStorage(ctx, geocloud.NewMessage(ctx.Param("id")))
	if err != nil {
		log.Err(err).Msg("unable to get storage")
		ctx.AbortWithStatus(statusCode)
		return
	}

	volume, err := a.os.GetObject(storage)
	if err != nil {
		log.Err(err).Msg("unable to get object")
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	b, contentType, statusCode, err := a.getVolumeContent(ctx, volume)
	if err != nil {
		log.Err(err).Msg("unable to get volume content")
		ctx.AbortWithStatus(statusCode)
		return
	}

	ctx.Data(http.StatusOK, contentType, b)
}

func (a *API) createBufferJobHandler(ctx *gin.Context) {
	job, statusCode, err := a.createJob(ctx, geocloud.TaskTypeBuffer)
	if err != nil {
		log.Err(err).Msg("unable to create job")
		ctx.AbortWithStatus(statusCode)
		return
	}

	ctx.JSON(http.StatusOK, job)
}

func (a *API) createFilterJobHandler(ctx *gin.Context) {
	job, statusCode, err := a.createJob(ctx, geocloud.TaskTypeFilter)
	if err != nil {
		log.Err(err).Msg("unable to create job")
		ctx.AbortWithStatus(statusCode)
		return
	}

	ctx.JSON(http.StatusOK, job)
}

func (a *API) createReprojectJobHandler(ctx *gin.Context) {
	job, statusCode, err := a.createJob(ctx, geocloud.TaskTypeReproject)
	if err != nil {
		log.Err(err).Msg("unable to create job")
		ctx.AbortWithStatus(statusCode)
		return
	}

	ctx.JSON(http.StatusOK, job)
}

func (a *API) createRemoveBadGeometryJobHandler(ctx *gin.Context) {
	job, statusCode, err := a.createJob(ctx, geocloud.TaskTypeRemoveBadGeometry)
	if err != nil {
		log.Err(err).Msg("unable to create job")
		ctx.AbortWithStatus(statusCode)
		return
	}

	ctx.JSON(http.StatusOK, job)
}

func (a *API) createVectorLookupJobHandler(ctx *gin.Context) {
	job, statusCode, err := a.createJob(ctx, geocloud.TaskTypeVectorLookup)
	if err != nil {
		log.Err(err).Msg("unable to create job")
		ctx.AbortWithStatus(statusCode)
		return
	}

	ctx.JSON(http.StatusOK, job)
}
