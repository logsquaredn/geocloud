package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
	"github.com/rs/zerolog/log"
)

func buildJobArgs(ctx *gin.Context, taskParams []string) []string {
	jobArgs := make([]string, len(taskParams))
	for i, param := range taskParams {
		jobArgs[i] = ctx.Query(param)
	}
	return jobArgs
}

type BufferParams struct {
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
// @Param distance query integer true "Buffer distance"
// @Param segmentCount query integer true "Segment count"
// @Success 200 {object} geocloud.CreateResponse
// @Failure 400 {object} geocloud.ErrorResponse
// @Failure 500 {object} geocloud.ErrorResponse
// @Router /create/buffer [post]
func (a *API) createBuffer(ctx *gin.Context) {
	var p BufferParams
	if err := ctx.ShouldBindQuery(&p); err != nil {
		log.Err(err).Msg("/createBuffer invalid query parameter(s)")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid query parameter(s): %s", err.Error())})
		return
	}

	a.create(ctx, "buffer")
}

// @Summary Create a remove bad geometry job
// @Description <b><u>Create a remove bad geometry job</u></b>
// @Description &emsp; - Pass the geospatial data to be processed in the request body
// @Tags createRemovebadgeometry
// @Accept application/json, application/zip
// @Produce application/json
// @Success 200 {object} geocloud.CreateResponse
// @Failure 400 {object} geocloud.ErrorResponse
// @Failure 500 {object} geocloud.ErrorResponse
// @Router /create/removebadgeometry [post]
func (a *API) createRemovebadgeometry(ctx *gin.Context) {
	a.create(ctx, "removebadgeometry")
}

type ReprojectParams struct {
	TargetProjection int `form:"targetProjection"`
}

// @Summary Create a reproject job
// @Description <b><u>Create a reproject job</u></b>
// @Description &emsp; - Pass the geospatial data to be processed in the request body
// @Tags createReproject
// @Accept application/json, application/zip
// @Produce application/json
// @Param targetProjection query integer true "Target projection EPSG"
// @Success 200 {object} geocloud.CreateResponse
// @Failure 400 {object} geocloud.ErrorResponse
// @Failure 500 {object} geocloud.ErrorResponse
// @Router /create/reproject [post]
func (a *API) createReproject(ctx *gin.Context) {
	var p ReprojectParams
	if err := ctx.ShouldBindQuery(&p); err != nil {
		log.Err(err).Msg("/createReproject invalid query parameter(s)")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid query parameter(s): %s", err.Error())})
		return
	}

	a.create(ctx, "reproject")
}

type FilterParams struct {
	FilterColumn string `json:"filterColumn"`
	FilterValue  string `json:"filterValue"`
}

// @Summary Create a filter job
// @Description <b><u>Create a filter job</u></b>
// @Description &emsp; - Pass the geospatial data to be processed in the request body
// @Tags createFilter
// @Accept application/json, application/zip
// @Produce application/json
// @Param filterColumn query string true "Column to filter on"
// @Param filterValue query string true "Value to filter on"
// @Success 200 {object} geocloud.CreateResponse
// @Failure 400 {object} geocloud.ErrorResponse
// @Failure 500 {object} geocloud.ErrorResponse
// @Router /create/filter [post]
func (a *API) createFilter(ctx *gin.Context) {
	var p FilterParams
	if err := ctx.ShouldBindQuery(&p); err != nil {
		log.Err(err).Msg("/createFilter invalid query parameter(s)")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid query parameter(s): %s", err.Error())})
		return
	}

	a.create(ctx, "filter")
}

type VectorlookupParams struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// @Summary Create a vector lookup job
// @Description <b><u>Create a vector lookup job</u></b>
// @Description &emsp; - Pass the geospatial data to be processed in the request body
// @Tags createVectorlookup
// @Accept application/json, application/zip
// @Produce application/json
// @Param longitude query number true "Longitude"
// @Param latitude query number true "Latitude"
// @Success 200 {object} geocloud.CreateResponse
// @Failure 400 {object} geocloud.ErrorResponse
// @Failure 500 {object} geocloud.ErrorResponse
// @Router /create/vectorlookup [post]
func (a *API) createVectorlookup(ctx *gin.Context) {
	var p VectorlookupParams
	if err := ctx.ShouldBindQuery(&p); err != nil {
		log.Err(err).Msg("/createVectorlookup invalid query parameter(s)")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid query parameter(s): %s", err.Error())})
		return
	}

	a.create(ctx, "vectorlookup")
}

func (a *API) create(ctx *gin.Context, whichTask string) {
	taskType, err := geocloud.TaskTypeFrom(whichTask)
	if err != nil {
		log.Err(err).Msgf("/create invalid task type requested: %s", whichTask)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid task type: %s", whichTask)})
		return
	}

	task, err := a.ds.GetTask(taskType)
	if err == sql.ErrNoRows {
		log.Err(err).Msgf("/create invalid task type requested: %s", taskType)
		ctx.JSON(http.StatusBadRequest, &geocloud.ErrorResponse{Error: fmt.Sprintf("invalid task type: %s", taskType)})
		return
	} else if err != nil {
		log.Err(err).Msgf("/create failed to query for params for type: %s", taskType)
		ctx.JSON(http.StatusInternalServerError, &geocloud.ErrorResponse{Error: fmt.Sprintf("failed to create job of type %s", taskType)})
		return
	}

	inputData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Err(err).Msgf("/create failed to read request body for type: %s", taskType)
		ctx.JSON(http.StatusInternalServerError, &geocloud.ErrorResponse{Error: "failed to read request body"})
		return
	}

	contentType := ctx.Request.Header.Get("Content-Type")
	var filename string
	if strings.Contains(contentType, "application/json") {
		if !isJSON(inputData) {
			log.Error().Msgf("/create received invalid json for type: %s", taskType)
			ctx.JSON(http.StatusBadRequest, &geocloud.ErrorResponse{Error: "request body must be valid JSON"})
			return
		}
		filename = "input.geojson"
	} else if strings.Contains(contentType, "application/zip") {
		if !isZIP(inputData) {
			log.Error().Msgf("/create received invalid zip for type: %s", taskType)
			ctx.JSON(http.StatusBadRequest, &geocloud.ErrorResponse{Error: "request body must be valid ZIP"})
			return
		}
		filename = "input.zip"
	} else {
		log.Error().Msgf("/create received invalid Content-Type: %s for type: %s", contentType, taskType)
		ctx.JSON(http.StatusBadRequest, &geocloud.ErrorResponse{Error: fmt.Sprintf("invalid Content-Type: %s", contentType)})
		return
	}

	customerID := getCustomerID(ctx)
	ist, err := a.ds.CreateStorage(&geocloud.Storage{
		CustomerID: customerID,
	})
	if err != nil {
		log.Err(err).Msg("/create failed to input storage for job")
		ctx.JSON(http.StatusInternalServerError, &geocloud.ErrorResponse{Error: fmt.Sprintf("/create failed to input storage for job")})
		return
	}

	job := &geocloud.Job{
		TaskType:   task.Type,
		Args:       buildJobArgs(ctx, task.Params),
		CustomerID: customerID,
		InputID:    ist.ID,
	}
	if job, err = a.ds.CreateJob(job); err != nil {
		log.Err(err).Msgf("/create failed to create job of type: %s", taskType)
		ctx.JSON(http.StatusInternalServerError, &geocloud.ErrorResponse{Error: fmt.Sprintf("failed to create job of type %s", taskType)})
		return
	}

	vol := geocloud.NewBytesVolume(filename, inputData)
	if err = a.os.PutObject(ist, vol); err != nil {
		log.Err(err).Msgf("/create failed to write data to objectstore for id: %s", job.GetID())
		ctx.JSON(http.StatusInternalServerError, &geocloud.ErrorResponse{Error: fmt.Sprintf("failed to create job of type %s", taskType)})
		job.Err = err
		job.Status = geocloud.Error
		a.ds.UpdateJob(job)
		return
	}

	if err = a.mq.Send(job); err != nil {
		log.Err(err).Msgf("/create failed to send message to messagequeue for id: %s", job.GetID())
		ctx.JSON(http.StatusInternalServerError, &geocloud.ErrorResponse{Error: fmt.Sprintf("failed send message for id %s with type %s", job.GetID(), taskType)})
		return
	}

	ctx.JSON(http.StatusOK, &geocloud.CreateResponse{ID: job.GetID()})
}

// @Summary Get status of a job
// @Description
// @Tags status
// @Produce json
// @Param id query string true "Job ID"
// @Success 200 {object} geocloud.StatusResponse
// @Failure 400 {object} geocloud.ErrorResponse
// @Failure 404 {object} geocloud.ErrorResponse
// @Failure 500 {object} geocloud.ErrorResponse
// @Router /status [get]
func (a *API) status(ctx *gin.Context) {
	id := ctx.Query("id")
	if len(id) < 1 {
		log.Error().Msg("/status query parameter 'id' not passed or empty")
		ctx.JSON(http.StatusBadRequest, &geocloud.ErrorResponse{Error: "query parameter 'id' required"})
		return
	}

	m := geocloud.NewMessage(id)
	job, err := a.ds.GetJob(m)
	if err == sql.ErrNoRows {
		log.Err(err).Msgf("/status got 0 results querying for id: %s", id)
		ctx.JSON(http.StatusNotFound, &geocloud.ErrorResponse{Error: "query parameter 'id' must be a valid job ID"})
		return
	} else if err != nil {
		log.Err(err).Msgf("/status failed to query for status for id: %s", id)
		ctx.JSON(http.StatusInternalServerError, &geocloud.ErrorResponse{Error: fmt.Sprintf("failed to get status for id: %s", id)})
		return
	}

	sr := &geocloud.StatusResponse{Status: job.Status.String()}
	if job.Status == geocloud.Error {
		sr.Error = job.Err.Error()
	}

	ctx.JSON(http.StatusOK, sr)
}

// @Summary Download geojson result of job
// @Description Results are downloadable as geojson or zip. The zip will contain the files that comprise an ESRI shapefile.
// @Tags result
// @Produce application/json, application/zip
// @Param id query string true "Job ID"
// @Success 200
// @Failure 400 {object} geocloud.ErrorResponse
// @Failure 404 {object} geocloud.ErrorResponse
// @Failure 500 {object} geocloud.ErrorResponse
// @Router /result [get]
func (a *API) result(ctx *gin.Context) {
	id := ctx.Query("id")
	if len(id) < 1 {
		log.Error().Msg("/result query parameter 'id' not passed or empty")
		ctx.JSON(http.StatusBadRequest, &geocloud.ErrorResponse{Error: "query parameter 'id' required"})
		return
	}

	m := geocloud.NewMessage(id)
	job, err := a.ds.GetJob(m)
	if err == sql.ErrNoRows {
		log.Err(err).Msgf("/result got 0 results querying for id: %s", id)
		ctx.JSON(http.StatusNotFound, &geocloud.ErrorResponse{Error: fmt.Sprintf("id: %s not found", id)})
		return
	} else if err != nil {
		log.Err(err).Msgf("/result failed to query for status for id: %s", id)
		ctx.JSON(http.StatusInternalServerError, &geocloud.ErrorResponse{Error: fmt.Sprintf("failed to find results for id: %s", id)})
		return
	} else if job.Status != geocloud.Complete {
		log.Error().Msgf("/result results requested but id: %s is of status: %s", id, job.Status)
		ctx.JSON(http.StatusBadRequest, &geocloud.ErrorResponse{Error: fmt.Sprintf("id: %s not complete. status: %s", id, job.Status)})
		return
	}

	vol, err := a.os.GetObject(geocloud.NewMessage(job.OutputID))
	if err != nil {
		log.Error().Msgf("/result failed to get result from s3 for id: %s", id)
		ctx.JSON(http.StatusInternalServerError, &geocloud.ErrorResponse{Error: fmt.Sprintf("failed to find results for id: %s", id)})
		return
	}

	wantZip := false
	if strings.Contains(ctx.Request.Header.Get("Accept"), "application/zip") {
		wantZip = true
	}
	var buf []byte
	err = vol.Walk(func(_ string, f geocloud.File, e error) error {
		if e != nil {
			return e
		}
		if wantZip && filepath.Ext(f.Name()) == ".zip" {
			buf = make([]byte, f.Size())
			_, e = f.Read(buf)
			return e
		} else if !wantZip && filepath.Ext(f.Name()) == ".geojson" {
			buf = make([]byte, f.Size())
			_, e = f.Read(buf)
			return e
		}

		return nil
	})
	if err != nil {
		log.Err(err).Msgf("/result failed to download result from s3 for id: %s", id)
		ctx.JSON(http.StatusInternalServerError, &geocloud.ErrorResponse{Error: fmt.Sprintf("failed to get results for id: %s", id)})
		return
	} else if len(buf) < 1 {
		log.Error().Msgf("/result downloaded no data from s3 for id: %s", id)
		ctx.JSON(http.StatusInternalServerError, &geocloud.ErrorResponse{Error: fmt.Sprintf("got no results for id: %s", id)})
		return
	}

	if wantZip {
		ctx.Data(http.StatusOK, "application/zip", buf)
	} else {
		var js map[string]interface{}
		err = json.Unmarshal(buf, &js)
		if err != nil || js == nil {
			log.Err(err).Msgf("/result failed to convert result to valid json for id: %s", id)
			ctx.JSON(http.StatusInternalServerError, &geocloud.ErrorResponse{Error: fmt.Sprintf("failed to create json result for id: %s", id)})
			return
		}

		ctx.JSON(http.StatusOK, js)
	}
}