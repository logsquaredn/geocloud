package api

import (
	"database/sql"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/pb"
)

// @Security     ApiKeyAuth
// @Summary      Get a list of jobs
// @Description  Get a list of jobs based on namespace
// @Tags         Job
// @Produce      application/json
// @Param        offset  query     int  false  "Offset of jobs to return"
// @Param        limit   query     int  false  "Limit of jobs to return"
// @Success      200     {object}  []rototiller.Job
// @Failure      401     {object}  rototiller.Error
// @Failure      500     {object}  rototiller.Error
// @Router       /api/v1/jobs [get].
func (a *Handler) listJobHandler(ctx *gin.Context) {
	q := &listQuery{}
	if err := ctx.BindQuery(q); err != nil {
		a.err(ctx, err)
		return
	}

	namespace, err := a.getNamespaceFromContext(ctx)
	if err != nil {
		a.err(ctx, err)
		return
	}

	jobs, err := a.Datastore.GetJobs(namespace, q.Offset, q.Limit)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		jobs = []*pb.Job{}
	case err != nil:
		a.err(ctx, err)
		return
	case jobs == nil:
		jobs = []*pb.Job{}
	}

	ctx.JSON(http.StatusOK, jobs)
}

// @Security     ApiKeyAuth
// @Summary      Get a job
// @Description  Get the metadata of a job. This can be used as a way to track job status
// @Tags         Job
// @Produce      application/json
// @Param        id   path      string  true  "Job ID"
// @Success      200  {object}  rototiller.Job
// @Failure      401  {object}  rototiller.Error
// @Failure      403  {object}  rototiller.Error
// @Failure      404  {object}  rototiller.Error
// @Failure      500  {object}  rototiller.Error
// @Router       /api/v1/jobs/{id} [get].
func (a *Handler) getJobHandler(ctx *gin.Context) {
	job, err := a.getJob(ctx, ctx.Param("job"))
	if err != nil {
		a.err(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, job)
}

// @Security     ApiKeyAuth
// @Summary      Get a job's tasks
// @Description  Get the metadata of a job's tasks
// @Tags         Task
// @Produce      application/json
// @Param        id   path      string  true  "Job ID"
// @Success      200  {object}  rototiller.Task
// @Failure      401  {object}  rototiller.Error
// @Failure      403  {object}  rototiller.Error
// @Failure      404  {object}  rototiller.Error
// @Failure      500  {object}  rototiller.Error
// @Router       /api/v1/jobs/{id}/tasks [get].
func (a *Handler) getJobTasksHandler(ctx *gin.Context) {
	job, err := a.getJob(ctx, ctx.Param("job"))
	if err != nil {
		a.err(ctx, err)
		return
	}

	tasks, err := a.getTasksFromJobSteps(job)
	if err != nil {
		a.err(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, tasks)
}

// @Security     ApiKeyAuth
// @Summary      Get a job's input
// @Description  Get the metadata of a job's input
// @Tags         Storage
// @Produce      application/json
// @Param        id   path      string  true  "Job ID"
// @Success      200  {object}  rototiller.Storage
// @Failure      401  {object}  rototiller.Error
// @Failure      403  {object}  rototiller.Error
// @Failure      404  {object}  rototiller.Error
// @Failure      500  {object}  rototiller.Error
// @Router       /api/v1/jobs/{id}/storages/input [get].
func (a *Handler) getJobInputHandler(ctx *gin.Context) {
	storage, err := a.getJobInputStorage(ctx, ctx.Param("job"))
	if err != nil {
		a.err(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, storage)
}

// @Security     ApiKeyAuth
// @Summary      Get a job's input content
// @Description  Gets the content of a job's input
// @Tags         Content
// @Produce      application/json, application/zip
// @Param        Accept  header  string  false  "Request results as a Zip or JSON. Default Zip"
// @Param        id      path    string  true   "Job ID"
// @Success      200
// @Failure      401  {object}  rototiller.Error
// @Failure      403  {object}  rototiller.Error
// @Failure      404  {object}  rototiller.Error
// @Failure      500  {object}  rototiller.Error
// @Router       /api/v1/jobs/{id}/storages/input/content [get].
func (a *Handler) getJobInputContentHandler(ctx *gin.Context) {
	storage, err := a.getJobInputStorage(ctx, ctx.Param("job"))
	if err != nil {
		a.err(ctx, err)
		return
	}

	volume, err := a.Blobstore.GetObject(ctx, storage.GetId())
	if err != nil {
		a.err(ctx, err)
		return
	}

	r, contentType, err := a.getVolumeContent(ctx.GetHeader("Accept"), volume)
	if err != nil {
		a.err(ctx, err)
		return
	}
	defer r.Close()

	ctx.Writer.Header().Add("Content-Type", contentType)
	_, _ = io.Copy(ctx.Writer, r)
}

// @Security     ApiKeyAuth
// @Summary      Get a job's output
// @Description  Get the metadata of a job's output
// @Tags         Storage
// @Produce      application/json
// @Param        id   path      string  true  "Job ID"
// @Success      200  {object}  rototiller.Storage
// @Failure      401  {object}  rototiller.Error
// @Failure      403  {object}  rototiller.Error
// @Failure      404  {object}  rototiller.Error
// @Failure      500  {object}  rototiller.Error
// @Router       /api/v1/jobs/{id}/storages/output [get].
func (a *Handler) getJobOutputHandler(ctx *gin.Context) {
	storage, err := a.getJobOutputStorage(ctx, ctx.Param("job"))
	if err != nil {
		a.err(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, storage)
}

// @Security     ApiKeyAuth
// @Summary      Get a job's output content
// @Description  Gets the content of a job's output
// @Tags         Content
// @Produce      application/json, application/zip
// @Param        Accept  header  string  false  "Request results as a Zip or JSON. Default Zip"
// @Param        id      path    string  true   "Job ID"
// @Success      200
// @Failure      401  {object}  rototiller.Error
// @Failure      403  {object}  rototiller.Error
// @Failure      404  {object}  rototiller.Error
// @Failure      500  {object}  rototiller.Error
// @Router       /api/v1/jobs/{id}/storages/output/content [get].
func (a *Handler) getJobOutputContentHandler(ctx *gin.Context) {
	storage, err := a.getJobOutputStorage(ctx, ctx.Param("job"))
	if err != nil {
		a.err(ctx, err)
		return
	}

	volume, err := a.Blobstore.GetObject(ctx, storage.GetId())
	if err != nil {
		a.err(ctx, err)
		return
	}

	r, contentType, err := a.getVolumeContent(ctx.GetHeader("Accept"), volume)
	if err != nil {
		a.err(ctx, err)
		return
	}
	defer r.Close()

	ctx.Writer.Header().Add("Content-Type", contentType)
	_, _ = io.Copy(ctx.Writer, r)
}

type bufferQuery struct {
	Distance     int `form:"buffer-distance" binding:"required"`
	SegmentCount int `form:"quadrant-segment-count" binding:"required"`
}

// @Security     ApiKeyAuth
// @Summary      Create a buffer job
// @Description  <b><u>Create a buffer job</u></b>
// @Description  &emsp; - Buffers every geometry by the given distance
// @Description
// @Description  &emsp; - For extra info: https://gdal.org/api/vector_c_pb.html#_CPPv412OGR_G_Buffer12OGRGeometryHdi
// @Description  &emsp; - Pass the geospatial data to be processed in the request body OR
// @Description  &emsp; - Pass the ID of an existing dataset with an empty request body
// @Description  &emsp; - This task accepts a ZIP containing a shapefile or GeoJSON input
// @Description  &emsp; - This task will automatically generate both GeoJSON and ZIP (shapfile) output
// @Tags         Job
// @Accept       application/json, application/zip
// @Produce      application/json
// @Param        Content-Type  header    string  false  "Required if passing geospatial data in request body"
// @Param        input                   query     string   false  "ID of existing dataset to use"
// @Param        input-of                query     string   false  "ID of existing job whose input dataset to use"
// @Param        output-of               query     string   false  "ID of existing job whose output dataset to use"
// @Param        buffer-distance         query     integer  true   "Buffer distance"
// @Param        quadrant-segment-count  query     integer  true   "Quadrant Segment count"
// @Success      200                     {object}  rototiller.Job
// @Failure      400                     {object}  rototiller.Error
// @Failure      401                     {object}  rototiller.Error
// @Failure      403                     {object}  rototiller.Error
// @Failure      500                     {object}  rototiller.Error
// @Router       /api/v1/jobs/buffer [post].
func (a *Handler) createBufferJobHandler(ctx *gin.Context) {
	if err := ctx.ShouldBindQuery(&bufferQuery{}); err != nil {
		a.err(ctx, pb.NewErr(err, http.StatusBadRequest))
		return
	}

	job, err := a.createJob(ctx, pb.TaskTypeBuffer)
	if err != nil {
		a.err(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, job)
}

type filterQuery struct {
	FilterColumn string `form:"filter-column" binding:"required"`
	FilterValue  string `form:"filter-value" binding:"required"`
}

// @Security     ApiKeyAuth
// @Summary      Create a filter job
// @Description  <b><u>Create a filter job</u></b>
// @Description  &emsp; - Drops features and their geometries that don't match the given filter
// @Description  &emsp; - Pass the geospatial data to be processed in the request body OR
// @Description  &emsp; - Pass the ID of an existing dataset with an empty request body
// @Description  &emsp; - This task accepts a ZIP containing a shapefile or GeoJSON input
// @Description  &emsp; - This task will automatically generate both GeoJSON and ZIP (shapfile) output
// @Tags         Job
// @Accept       application/json, application/zip
// @Produce      application/json
// @Param        Content-Type   header    string  false  "Required if passing geospatial data in request body"
// @Param        input          query     string  false  "ID of existing dataset to use"
// @Param        input-of       query     string  false  "ID of existing job whose input dataset to use"
// @Param        output-of      query     string  false  "ID of existing job whose output dataset to use"
// @Param        filter-column  query     string  true   "Column to filter on"
// @Param        filter-value   query     string  true   "Value to filter on"
// @Success      200            {object}  rototiller.Job
// @Failure      400            {object}  rototiller.Error
// @Failure      401            {object}  rototiller.Error
// @Failure      403            {object}  rototiller.Error
// @Failure      500            {object}  rototiller.Error
// @Router       /api/v1/jobs/filter [post].
func (a *Handler) createFilterJobHandler(ctx *gin.Context) {
	if err := ctx.ShouldBindQuery(&filterQuery{}); err != nil {
		a.err(ctx, pb.NewErr(err, http.StatusBadRequest))
		return
	}

	job, err := a.createJob(ctx, pb.TaskTypeFilter)
	if err != nil {
		a.err(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, job)
}

type reprojectQuery struct {
	TargetProjection int `form:"target-projection" binding:"required"`
}

// @Security     ApiKeyAuth
// @Summary      Create a reproject job
// @Description  <b><u>Create a reproject job</u></b>
// @Description  &emsp; - Reprojects all geometries to the given projection
// @Description  &emsp; - Pass the geospatial data to be processed in the request body OR
// @Description  &emsp; - Pass the ID of an existing dataset with an empty request body
// @Description  &emsp; - This task accepts a ZIP containing a shapefile or GeoJSON input
// @Description  &emsp; - This task will automatically generate both GeoJSON and ZIP (shapfile) output
// @Tags         Job
// @Accept       application/json, application/zip
// @Produce      application/json
// @Param        Content-Type       header    string   false  "Required if passing geospatial data in request body"
// @Param        input              query     string   false  "ID of existing dataset to use"
// @Param        input-of           query     string   false  "ID of existing job whose input dataset to use"
// @Param        output-of          query     string   false  "ID of existing job whose output dataset to use"
// @Param        target-projection  query     integer  true   "Target projection EPSG"
// @Success      200                {object}  rototiller.Job
// @Failure      400                {object}  rototiller.Error
// @Failure      401                {object}  rototiller.Error
// @Failure      403                {object}  rototiller.Error
// @Failure      500                {object}  rototiller.Error
// @Router       /api/v1/jobs/reproject [post].
func (a *Handler) createReprojectJobHandler(ctx *gin.Context) {
	if err := ctx.ShouldBindQuery(&reprojectQuery{}); err != nil {
		a.err(ctx, pb.NewErr(err, http.StatusBadRequest))
		return
	}

	job, err := a.createJob(ctx, pb.TaskTypeReproject)
	if err != nil {
		a.err(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, job)
}

// @Security     ApiKeyAuth
// @Summary      Create a remove bad geometry job
// @Description  <b><u>Create a remove bad geometry job</u></b>
// @Description  &emsp; - Drops geometries that are invalid
// @Description
// @Description  &emsp; - For extra info: https://gdal.org/api/vector_c_pb.html#_CPPv413OGR_G_IsValid12OGRGeometryH
// @Description  &emsp; - API Key is required either as a query parameter or a header
// @Description  &emsp; - Pass the geospatial data to be processed in the request body OR
// @Description  &emsp; - Pass the ID of an existing dataset with an empty request body
// @Description  &emsp; - This task accepts a ZIP containing a shapefile or GeoJSON input
// @Description  &emsp; - This task will automatically generate both GeoJSON and ZIP (shapfile) output
// @Tags         Job
// @Accept       application/json, application/zip
// @Produce      application/json
// @Param        Content-Type  header    string  false  "Required if passing geospatial data in request body"
// @Param        input         query     string  false  "ID of existing dataset to use"
// @Param        input-of      query     string  false  "ID of existing job whose input dataset to use"
// @Param        output-of     query     string  false  "ID of existing job whose output dataset to use"
// @Success      200           {object}  rototiller.Job
// @Failure      400           {object}  rototiller.Error
// @Failure      401           {object}  rototiller.Error
// @Failure      403           {object}  rototiller.Error
// @Failure      500           {object}  rototiller.Error
// @Router       /api/v1/jobs/removebadgeometry [post].
func (a *Handler) createRemoveBadGeometryJobHandler(ctx *gin.Context) {
	job, err := a.createJob(ctx, pb.TaskTypeRemoveBadGeometry)
	if err != nil {
		a.err(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, job)
}

type vectorLookupQuery struct {
	Attributes string  `form:"attributes" binding:"required"`
	Longitude  float64 `form:"longitude" binding:"required"`
	Latitude   float64 `form:"latitude" binding:"required"`
}

// @Security     ApiKeyAuth
// @Summary      Create a vector lookup job
// @Description  <b><u>Create a vector lookup job</u></b>
// @Description  &emsp; - Returns a list of attribute values of which the given point intersects
// @Description  &emsp; - Pass the geospatial data to be processed in the request body OR
// @Description  &emsp; - Pass the ID of an existing dataset with an empty request body
// @Description  &emsp; - This task accepts a ZIP containing a shapefile or GeoJSON input
// @Description  &emsp; - This task will generate JSON output
// @Tags         Job
// @Accept       application/json, application/zip
// @Produce      application/json
// @Param        Content-Type  header    string  false  "Required if passing geospatial data in request body"
// @Param        input         query     string  false  "ID of existing dataset to use"
// @Param        input-of      query     string  false  "ID of existing job whose input dataset to use"
// @Param        output-of     query     string  false  "ID of existing job whose output dataset to use"
// @Param        attributes    query     string  true   "Comma separated list of attributes"
// @Param        longitude     query     number  true   "Longitude"
// @Param        latitude      query     number  true   "Latitude"
// @Success      200           {object}  rototiller.Job
// @Failure      400           {object}  rototiller.Error
// @Failure      401           {object}  rototiller.Error
// @Failure      403           {object}  rototiller.Error
// @Failure      500           {object}  rototiller.Error
// @Router       /api/v1/jobs/vectorlookup [post].
func (a *Handler) createVectorLookupJobHandler(ctx *gin.Context) {
	if err := ctx.ShouldBindQuery(&vectorLookupQuery{}); err != nil {
		a.err(ctx, pb.NewErr(err, http.StatusBadRequest))
		return
	}

	job, err := a.createJob(ctx, pb.TaskTypeVectorLookup)
	if err != nil {
		a.err(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, job)
}

type rasterLookupQuery struct {
	Bands     string  `form:"bands" binding:"required"`
	Longitude float64 `form:"longitude" binding:"required"`
	Latitude  float64 `form:"latitude" binding:"required"`
}

// @Security     ApiKeyAuth
// @Summary      Create a raster lookup job
// @Description  <b><u>Create a raster lookup job</u></b>
// @Description  &emsp; - Returns the value of each requested band of which the given point intersects
// @Description  &emsp; - Pass the geospatial data to be processed in the request body OR
// @Description  &emsp; - Pass the ID of an existing dataset with an empty request body
// @Description  &emsp; - This task accepts a ZIP containing a single TIF file. Valid extensions are: tif, tiff, geotif, geotiff
// @Description  &emsp; - This task will generate JSON output
// @Tags         Job
// @Accept       application/json, application/zip
// @Produce      application/json
// @Param        Content-Type            header    string   false  "Required if passing geospatial data in request body"
// @Param        input         query     string  false  "ID of existing dataset to use"
// @Param        input-of      query     string  false  "ID of existing job whose input dataset to use"
// @Param        output-of     query     string  false  "ID of existing job whose output dataset to use"
// @Param        bands         query     string  true   "Comma separated list of bands"
// @Param        longitude     query     number  true   "Longitude"
// @Param        latitude      query     number  true   "Latitude"
// @Success      200           {object}  rototiller.Job
// @Failure      400           {object}  rototiller.Error
// @Failure      401           {object}  rototiller.Error
// @Failure      403           {object}  rototiller.Error
// @Failure      500           {object}  rototiller.Error
// @Router       /api/v1/jobs/rasterlookup [post].
func (a *Handler) createRasterLookupJobHandler(ctx *gin.Context) {
	if err := ctx.ShouldBindQuery(&rasterLookupQuery{}); err != nil {
		a.err(ctx, pb.NewErr(err, http.StatusBadRequest))
		return
	}

	job, err := a.createJob(ctx, pb.TaskTypeRasterLookup)
	if err != nil {
		a.err(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, job)
}

type polygonVectorLookupQuery struct {
	Attributes string `form:"attributes" binding:"required"`
	Polygon    string `form:"polygon" binding:"required"`
}

// @Security     ApiKeyAuth
// @Summary      Create a polygon vector lookup job
// @Description  <b><u>Create a polygon vector lookup job</u></b>
// @Description  &emsp; - Returns a list of attribute values of which the given polygon intersects
// @Description  &emsp; - Pass the geospatial data to be processed in the request body OR
// @Description  &emsp; - Pass the ID of an existing dataset with an empty request body
// @Description  &emsp; - This task accepts a ZIP containing a shapefile or GeoJSON input
// @Description  &emsp; - This task will generate JSON output
// @Tags         Job
// @Accept       application/json, application/zip
// @Produce      application/json
// @Param        Content-Type  header    string  false  "Required if passing geospatial data in request body"
// @Param        input         query     string  false  "ID of existing dataset to use"
// @Param        input-of      query     string  false  "ID of existing job whose input dataset to use"
// @Param        output-of     query     string  false  "ID of existing job whose output dataset to use"
// @Param        attributes    query     string  true   "Comma separated list of attributes"
// @Param        polygon       query     string  true   "Polygon in WKT format"
// @Success      200           {object}  rototiller.Job
// @Failure      400           {object}  rototiller.Error
// @Failure      401           {object}  rototiller.Error
// @Failure      403           {object}  rototiller.Error
// @Failure      500           {object}  rototiller.Error
// @Router       /api/v1/jobs/polygonvectorlookup [post].
func (a *Handler) createPolygonVectorLookupJobHandler(ctx *gin.Context) {
	if err := ctx.ShouldBindQuery(&polygonVectorLookupQuery{}); err != nil {
		a.err(ctx, pb.NewErr(err, http.StatusBadRequest))
		return
	}

	job, err := a.createJob(ctx, pb.TaskTypePolygonVectorLookup)
	if err != nil {
		a.err(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, job)
}

// type taskChainQuery struct {
// 	InputID string `form:"attributes" binding:"required"`
// }

// func (a *Handler) createTaskChainJobHandler(ctx *gin.Context) {
// 	if err := ctx.ShouldBindQuery(&taskChainQuery{}); err != nil {
// 		a.err(ctx, api.NewErr(err, http.StatusBadRequest))
// 		return
// 	}

// }
