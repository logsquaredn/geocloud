package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
)

// @Summary Get a list of jobs
// @Description
// @Tags
// @Produce application/json
// @Success 200 {object} []geocloud.Job
// @Failure 401 {object} geocloud.Error
// @Failure 404 {object} geocloud.Error
// @Failure 500 {object} geocloud.Error
// @Router /job [get]
func (a *API) listJobHandler(ctx *gin.Context) {
	jobs, err := a.ds.GetCustomerJobs(a.getAssumedCustomer(ctx))
	switch {
	case err == sql.ErrNoRows:
		a.err(ctx, http.StatusNotFound, err)
	case err != nil:
		a.err(ctx, http.StatusInsufficientStorage, err)
	}

	ctx.JSON(http.StatusOK, jobs)
}

// @Summary Get a job
// @Description
// @Tags
// @Produce application/json
// @Param id path string true "Job ID"
// @Success 200 {object} geocloud.Job
// @Failure 401 {object} geocloud.Error
// @Failure 403 {object} geocloud.Error
// @Failure 404 {object} geocloud.Error
// @Failure 500 {object} geocloud.Error
// @Router /job/{id} [get]
func (a *API) getJobHandler(ctx *gin.Context) {
	job, statusCode, err := a.getJob(ctx, geocloud.NewMessage(ctx.Param("id")))
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	ctx.JSON(http.StatusOK, job)
}

// @Summary Get a job's task
// @Description
// @Tags
// @Produce application/json
// @Param id path string true "Job ID"
// @Success 200 {object} geocloud.Task
// @Failure 401 {object} geocloud.Error
// @Failure 403 {object} geocloud.Error
// @Failure 404 {object} geocloud.Error
// @Failure 500 {object} geocloud.Error
// @Router /job/{id}/task [get]
func (a *API) getJobTaskHandler(ctx *gin.Context) {
	job, statusCode, err := a.getJob(ctx, geocloud.NewMessage(ctx.Param("id")))
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	task, statusCode, err := a.getTaskType(ctx, job.TaskType)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	ctx.JSON(http.StatusOK, task)
}

// @Summary Get a job's input
// @Description
// @Tags
// @Produce application/json
// @Param id path string true "Job ID"
// @Success 200 {object} geocloud.Storage
// @Failure 401 {object} geocloud.Error
// @Failure 403 {object} geocloud.Error
// @Failure 404 {object} geocloud.Error
// @Failure 500 {object} geocloud.Error
// @Router /job/{id}/input [get]
func (a *API) getJobInputHandler(ctx *gin.Context) {
	storage, statusCode, err := a.getJobInputStorage(ctx, geocloud.NewMessage(ctx.Param("id")))
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	ctx.JSON(http.StatusOK, storage)
}

// @Summary Get a job's input content
// @Description
// @Tags
// @Produce application/json, application/zip
// @Param id path string true "Job ID"
// @Success 200
// @Failure 401 {object} geocloud.Error
// @Failure 403 {object} geocloud.Error
// @Failure 404 {object} geocloud.Error
// @Failure 500 {object} geocloud.Error
// @Router /job/{id}/input/content [get]
func (a *API) getJobInputContentHandler(ctx *gin.Context) {
	storage, statusCode, err := a.getJobInputStorage(ctx, geocloud.NewMessage(ctx.Param("id")))
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	volume, err := a.os.GetObject(storage)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	b, contentType, statusCode, err := a.getVolumeContent(ctx, volume)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	ctx.Data(http.StatusOK, contentType, b)
}

// @Summary Get a job's output
// @Description
// @Tags
// @Produce application/json
// @Param id path string true "Job ID"
// @Success 200 {object} geocloud.Storage
// @Failure 401 {object} geocloud.Error
// @Failure 403 {object} geocloud.Error
// @Failure 404 {object} geocloud.Error
// @Failure 500 {object} geocloud.Error
// @Router /job/{id}/output [get]
func (a *API) getJobOutputHandler(ctx *gin.Context) {
	storage, statusCode, err := a.getJobOutputStorage(ctx, geocloud.NewMessage(ctx.Param("id")))
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	ctx.JSON(http.StatusOK, storage)
}

// @Summary Get a job's output content
// @Description
// @Tags
// @Produce application/json, application/zip
// @Param id path string true "Job ID"
// @Success 200
// @Failure 401 {object} geocloud.Error
// @Failure 403 {object} geocloud.Error
// @Failure 404 {object} geocloud.Error
// @Failure 500 {object} geocloud.Error
// @Router /job/{id}/output/content [get]
func (a *API) getJobOutputContentHandler(ctx *gin.Context) {
	storage, statusCode, err := a.getJobOutputStorage(ctx, geocloud.NewMessage(ctx.Param("id")))
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	volume, err := a.os.GetObject(storage)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	b, contentType, statusCode, err := a.getVolumeContent(ctx, volume)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	ctx.Data(http.StatusOK, contentType, b)
}

type bufferQuery struct {
	Distance     int `form:"distance"`
	SegmentCount int `form:"segmentCount"`
}

// @Summary Create a buffer job
// @Description <b><u>Create a buffer job</u></b>
// @Description &emsp; - For extra info: https://gdal.org/api/vector_c_api.html#_CPPv412OGR_G_Buffer12OGRGeometryHdi
// @Description &emsp; - Pass the geospatial data to be processed in the request body.
// @Tags createBuffer
// @Accept application/json, application/zip
// @Produce application/json
// @Param input_id query string false "Optional ID of existing storage to use"
// @Param distance query integer true "Buffer distance"
// @Param segmentCount query integer true "Segment count"
// @Success 200 {object} geocloud.Job
// @Failure 400 {object} geocloud.Error
// @Failure 401 {object} geocloud.Error
// @Failure 403 {object} geocloud.Error
// @Failure 500 {object} geocloud.Error
// @Router /job/buffer [post]
func (a *API) createBufferJobHandler(ctx *gin.Context) {
	if err := ctx.ShouldBindQuery(&bufferQuery{}); err != nil {
		a.err(ctx, http.StatusBadRequest, err)
		return
	}

	job, statusCode, err := a.createJob(ctx, geocloud.TaskTypeBuffer)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	ctx.JSON(http.StatusOK, job)
}

type filterQuery struct {
	FilterColumn string `form:"filterColumn"`
	FilterValue  string `form:"filterValue"`
}

// @Summary Create a filter job
// @Description <b><u>Create a filter job</u></b>
// @Description &emsp; - Pass the geospatial data to be processed in the request body
// @Tags createFilter
// @Accept application/json, application/zip
// @Produce application/json
// @Param input_id query string false "Optional ID of existing storage to use"
// @Param filterColumn query string true "Column to filter on"
// @Param filterValue query string true "Value to filter on"
// @Success 200 {object} geocloud.Job
// @Failure 400 {object} geocloud.Error
// @Failure 401 {object} geocloud.Error
// @Failure 403 {object} geocloud.Error
// @Failure 500 {object} geocloud.Error
// @Router /job/filter [post]
func (a *API) createFilterJobHandler(ctx *gin.Context) {
	if err := ctx.ShouldBindQuery(&filterQuery{}); err != nil {
		a.err(ctx, http.StatusBadRequest, err)
		return
	}

	job, statusCode, err := a.createJob(ctx, geocloud.TaskTypeFilter)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	ctx.JSON(http.StatusOK, job)
}

type reprojectQuery struct {
	TargetProjection int `form:"targetProjection"`
}

// @Summary Create a reproject job
// @Description <b><u>Create a reproject job</u></b>
// @Description &emsp; - Pass the geospatial data to be processed in the request body
// @Tags createReproject
// @Accept application/json, application/zip
// @Produce application/json
// @Param input_id query string false "Optional ID of existing storage to use"
// @Param targetProjection query integer true "Target projection EPSG"
// @Success 200 {object} geocloud.Job
// @Failure 400 {object} geocloud.Error
// @Failure 401 {object} geocloud.Error
// @Failure 403 {object} geocloud.Error
// @Failure 500 {object} geocloud.Error
// @Router /job/reproject [post]
func (a *API) createReprojectJobHandler(ctx *gin.Context) {
	if err := ctx.ShouldBindQuery(&reprojectQuery{}); err != nil {
		a.err(ctx, http.StatusBadRequest, err)
		return
	}

	job, statusCode, err := a.createJob(ctx, geocloud.TaskTypeReproject)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	ctx.JSON(http.StatusOK, job)
}

// @Summary Create a remove bad geometry job
// @Description <b><u>Create a remove bad geometry job</u></b>
// @Description &emsp; - Pass the geospatial data to be processed in the request body
// @Tags createRemovebadgeometry
// @Accept application/json, application/zip
// @Produce application/json
// @Param input_id query string false "Optional ID of existing storage to use"
// @Success 200 {object} geocloud.Job
// @Failure 400 {object} geocloud.Error
// @Failure 401 {object} geocloud.Error
// @Failure 403 {object} geocloud.Error
// @Failure 500 {object} geocloud.Error
// @Router /job/removebadgeometry [post]
func (a *API) createRemoveBadGeometryJobHandler(ctx *gin.Context) {
	job, statusCode, err := a.createJob(ctx, geocloud.TaskTypeRemoveBadGeometry)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	ctx.JSON(http.StatusOK, job)
}

type vectorlookupQuery struct {
	Longitude float64 `form:"longitude"`
	Latitude  float64 `form:"latitude"`
}

// @Summary Create a vector lookup job
// @Description <b><u>Create a vector lookup job</u></b>
// @Description &emsp; - Pass the geospatial data to be processed in the request body
// @Tags createVectorlookup
// @Accept application/json, application/zip
// @Produce application/json
// @Param input_id query string false "Optional ID of existing storage to use"
// @Param longitude query number true "Longitude"
// @Param latitude query number true "Latitude"
// @Success 200 {object} geocloud.Job
// @Failure 400 {object} geocloud.Error
// @Failure 401 {object} geocloud.Error
// @Failure 403 {object} geocloud.Error
// @Failure 500 {object} geocloud.Error
// @Router /job/vectorlookup [post]
func (a *API) createVectorLookupJobHandler(ctx *gin.Context) {
	if err := ctx.ShouldBindQuery(&vectorlookupQuery{}); err != nil {
		a.err(ctx, http.StatusBadRequest, err)
		return
	}

	job, statusCode, err := a.createJob(ctx, geocloud.TaskTypeVectorLookup)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	ctx.JSON(http.StatusOK, job)
}
