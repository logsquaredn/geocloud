package router

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (h *Router) createJob(jobType string) (id string, err error) {
	id = uuid.New().String()
	return id, h.das.InsertNewJob(id, jobType)
}

func validateParamsPassed(ctx *gin.Context, taskParams []string) (missingParams []string) {
	missingParams = []string{}
	for _, param := range taskParams {
		if len(ctx.Query(param)) < 1 {
			missingParams = append(missingParams, param)
		}
	}

	return missingParams
}

func (h *Router) create(ctx *gin.Context) {
	taskType := ctx.Param("type")
	taskParams, err := h.das.GetTaskParamsByTaskType(taskType)
	if err == sql.ErrNoRows {
		log.Error().Msgf("/create invalid task type requested: %s", taskType)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid task type: %s", taskType)})
		return
	} else if err != nil {
		log.Err(err).Msgf("/create failed to query for params for type: %s", taskType)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create job of type %s", taskType)})
		return
	}

	missingParams := validateParamsPassed(ctx, taskParams)
	if len(missingParams) > 0 {
		log.Error().Msgf("/create missing paramters: %v", missingParams)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("missing parameters: %v", missingParams)})
	}

	jsonData, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil || !isJSON(jsonData) {
		log.Err(err).Msgf("/create received invalid json for type: %s", taskType)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "request body must be valid JSON"})
		return
	}

	id, err := h.createJob(taskType)
	if err != nil {
		log.Err(err).Msgf("/create failed to create job of type: %s", taskType)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create job of type %s", taskType)})
		return
	}

	// TODO write jsonData to s3

	// TODO send SQS message

	ctx.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *Router) status(ctx *gin.Context) {
	id := ctx.Query("id")
	if len(id) < 1 {
		log.Error().Msg("/status query paramter 'id' not passed or empty")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'id' required"})
		return
	}

	status, err := h.das.GetJobStatusByJobId(id)
	if err == sql.ErrNoRows {
		log.Err(err).Msgf("/status got 0 results querying for id: %s", id)
		ctx.Status(http.StatusNotFound)
		return
	} else if err != nil {
		log.Err(err).Msgf("/status failed to query for status for id: %s", id)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed get get status for id: %s", id)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": status})
}
