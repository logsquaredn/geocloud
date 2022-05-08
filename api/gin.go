package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
	"github.com/logsquaredn/geocloud/docs"
	"github.com/rs/zerolog/log"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

// @contact.name logsquaredn
// @contact.url https://logsquaredn.io
// @contact.email logsquaredn@gmail.com

// @license.name logsquaredn

type ginAPI struct {
	ds     geocloud.Datastore
	os     geocloud.Objectstore
	mq     geocloud.MessageRecipient
	router *gin.Engine
}

var _ geocloud.API = (*ginAPI)(nil)

func init() {
	docs.SwaggerInfo.Title = "Geocloud"
	docs.SwaggerInfo.Description = "Geocloud"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "geocloud.logsquaredn.io"
	docs.SwaggerInfo.BasePath = "/api/v1/job"
	docs.SwaggerInfo.Schemes = []string{"https"}
}

func NewServer(opts *GinOpts) (*ginAPI, error) {
	var (
		a = &ginAPI{
			ds:     opts.Datastore,
			os:     opts.Objectstore,
			mq:     opts.MessageQueue,
			router: gin.Default(),
		}
	)

	a.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	a.router.Use(a.middleware)

	v1Job := a.router.Group("/api/v1/job")
	{
		v1Job.POST("/create/:type", a.create)
		v1Job.GET("/status", a.status)
		v1Job.GET("/result", a.result)
	}

	return a, nil
}

func (a *ginAPI) Serve(l net.Listener) error {
	return a.router.RunListener(l)
}

const (
	apiKeyQueryParam = "api-key"
	apiKeyHeader     = "X-API-Key"
	apiKeyCookie     = apiKeyHeader
)

func (a *ginAPI) middleware(ctx *gin.Context) {
	apiKey := ctx.Query(apiKeyQueryParam)
	if apiKey == "" {
		apiKey = ctx.GetHeader(apiKeyHeader)
		if apiKey == "" {
			apiKey, _ = ctx.Cookie(apiKeyCookie)
		}
	}

	if _, err := a.ds.GetCustomer(apiKey); err != nil {
		if err == sql.ErrNoRows {
			log.Err(err).Msgf("query parameter '%s', header '%s' or cookie '%s' must be a valid API Key", apiKeyQueryParam, apiKeyHeader, apiKeyCookie)
			ctx.AbortWithStatusJSON(http.StatusForbidden, &geocloud.ErrorResponse{Error: fmt.Sprintf("query parameter '%s', header '%s' or cookie '%s' must be a valid API Key", apiKeyQueryParam, apiKeyHeader, apiKeyCookie)})
			return
		}

		log.Err(err).Msgf("failed to get user information")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, &geocloud.ErrorResponse{Error: "failed to get user information"})
		return
	}

	ctx.Next()
}

func validateParamsPassed(ctx *gin.Context, taskParams []string) (missingParams []string) {
	for _, param := range taskParams {
		if len(ctx.Query(param)) < 1 {
			missingParams = append(missingParams, param)
		}
	}

	return
}

func buildJobArgs(ctx *gin.Context, taskParams []string) []string {
	jobArgs := make([]string, len(taskParams))
	for i, param := range taskParams {
		jobArgs[i] = ctx.Query(param)
	}
	return jobArgs
}

// @Summary Create a job
// @Description <b><u>Create a job</u></b>
// @Description <b>1. {type}: buffer {params}: distance, quadSegCount</b>
// @Description &emsp; - For info: https://gdal.org/api/vector_c_api.html#_CPPv412OGR_G_Buffer12OGRGeometryHdi
// @Description <br>
// @Description <b>2. {type}: filter {params}: filterColumn, filterValue</b>
// @Description <br>
// @Description <b>3. {type}: reproject {params}: targetProjection</b>
// @Description &emsp; - targetProjection should be an EPSG code
// @Description <br>
// @Description <b>4. {type}: removebadgeometry</b>
// @Description &emsp; - For info: https://gdal.org/api/vector_c_api.html#_CPPv413OGR_G_IsValid12OGRGeometryH
// @Description <br>
// @Description Pass the geojson to be processed in the body.
// @Tags create
// @Accept json
// @Produce json
// @Param type path string true "Job Type"
// @Success 200 {object} geocloud.CreateResponse
// @Failure 400 {object} geocloud.ErrorResponse
// @Failure 500 {object} geocloud.ErrorResponse
// @Router /create/{type} [post]
func (a *ginAPI) create(ctx *gin.Context) {
	taskType, err := geocloud.TaskTypeFrom(ctx.Param("type"))
	if err != nil {
		log.Err(err).Msgf("/create invalid task type requested: %s", taskType)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid task type: %s", taskType)})
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

	missingParams := validateParamsPassed(ctx, task.Params)
	if len(missingParams) > 0 {
		log.Error().Msgf("/create missing parameters: %v", missingParams)
		ctx.JSON(http.StatusBadRequest, &geocloud.ErrorResponse{Error: fmt.Sprintf("missing parameters: %v", missingParams)})
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

	job := &geocloud.Job{
		TaskType:   task.Type,
		Args:       buildJobArgs(ctx, task.Params),
		CustomerID: ctx.GetHeader("customer-id"),
	}
	if job, err = a.ds.CreateJob(job); err != nil {
		log.Err(err).Msgf("/create failed to create job of type: %s", taskType)
		ctx.JSON(http.StatusInternalServerError, &geocloud.ErrorResponse{Error: fmt.Sprintf("failed to create job of type %s", taskType)})
		return
	}

	vol := &bytesVolume{
		reader: bytes.NewReader(inputData),
		name:   filename,
	}
	if err = a.os.PutInput(job, vol); err != nil {
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
func (a *ginAPI) status(ctx *gin.Context) {
	id := ctx.Query("id")
	if len(id) < 1 {
		log.Error().Msg("/status query parameter 'id' not passed or empty")
		ctx.JSON(http.StatusBadRequest, &geocloud.ErrorResponse{Error: "query parameter 'id' required"})
		return
	}

	m := &message{id: id}
	job, err := a.ds.GetJob(m)
	if err == sql.ErrNoRows {
		log.Err(err).Msgf("/status got 0 results querying for id: %s", id)
		// TODO add message
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
func (a *ginAPI) result(ctx *gin.Context) {
	id := ctx.Query("id")
	if len(id) < 1 {
		log.Error().Msg("/result query parameter 'id' not passed or empty")
		ctx.JSON(http.StatusBadRequest, &geocloud.ErrorResponse{Error: "query parameter 'id' required"})
		return
	}

	m := &message{id: id}
	job, err := a.ds.GetJob(m)
	if err == sql.ErrNoRows {
		log.Err(err).Msgf("/result got 0 results querying for id: %s", id)
		// TODO add message
		ctx.Status(http.StatusNotFound)
		return
	} else if err != nil {
		log.Err(err).Msgf("/result failed to query for status for id: %s", id)
		// TODO add message
		ctx.Status(http.StatusInternalServerError)
		return
	} else if job.Status != geocloud.Complete {
		log.Error().Msgf("/result results requested but id: %s is of status: %s", id, job.Status)
		// TODO add message
		ctx.Status(http.StatusBadRequest)
		return
	}

	vol, err := a.os.GetOutput(m)
	if err != nil {
		log.Error().Msgf("/result failed to get result from s3 for id: %s", id)
		// TODO add message
		ctx.Status(http.StatusInternalServerError)
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
		// TODO add message
		ctx.Status(http.StatusInternalServerError)
		return
	} else if len(buf) < 1 {
		log.Error().Msgf("/result downloaded no data from s3 for id: %s", id)
		// TODO add message
		ctx.Status(http.StatusInternalServerError)
		return
	}

	if wantZip {
		ctx.Data(http.StatusOK, "application/zip", buf)
	} else {
		var js map[string]interface{}
		json.Unmarshal(buf, &js)
		if js == nil {
			log.Error().Msgf("/result failed to convert result to valid json for id: %s", id)
			// TODO add message
			ctx.Status(http.StatusInternalServerError)
			return
		}

		ctx.JSON(http.StatusOK, js)
	}
}

func isJSON(jsBytes []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(jsBytes, &js) == nil
}

func isZIP(zipBytes []byte) bool {
	return http.DetectContentType(zipBytes) == "application/zip"
}
