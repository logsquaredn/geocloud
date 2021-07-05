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

func (h *Router) buffer(ctx *gin.Context) {
	distance := ctx.Query("distance")
	if len(distance) < 1 {
		log.Error().Msg("/buffer query paramter 'distance' not passed or empty")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'distance' required"})
		return
	}

	jsonData, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil || !isJSON(jsonData) {
		log.Err(err).Msg("/buffer received invalid JSON")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "request body must be valid json"})
		return
	}

	var jobType = "buffer"
	id, err := h.createJob(jobType)
	if err != nil {
		log.Err(err).Msgf("/buffer failed to create job with id: %s and type: %s", id, jobType)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create %s job", jobType)})
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
		log.Err(err).Msgf("/status failed to query row for id: %s", id)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": status})
}
