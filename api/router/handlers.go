package router

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (r *Router) createJob(jobType string) (id string, err error) {
	id = uuid.New().String()
	return id, r.das.InsertNewJob(id, jobType)
}

func validateParamsPassed(ctx *gin.Context, taskParams []string) (missingParams []string) {
	for _, param := range taskParams {
		if len(ctx.Query(param)) < 1 {
			missingParams = append(missingParams, param)
		}
	}

	return
}

func (r *Router) create(ctx *gin.Context) {
	taskType := ctx.Param("type")
	if len(taskType) < 1 {
		log.Error().Msg("/create query paramter 'type' not passed or empty")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'type' required"})
		return
	}

	taskParams, err := r.das.GetTaskParamsByTaskType(taskType)
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

	id, err := r.createJob(taskType)
	if err != nil {
		log.Err(err).Msgf("/create failed to create job of type: %s", taskType)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create job of type %s", taskType)})
		return
	}

	_, err = r.oas.PutJobInput(id, bytes.NewReader(jsonData), "geojson")
	if err != nil {
		log.Err(err).Msgf("/create failed to write data to s3 for id: %s", id)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create job of type %s", taskType)})
		return
	}

	queueName, err := r.das.GetQueueNameByTaskType(taskType)
	if err != nil {
		log.Err(err).Msgf("/create failed to get queue name for task type: %s", taskType)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create job of type %s", taskType)})
		return
	}

	queueUrlOutput, err := r.sqs.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})
	if err != nil {
		log.Err(err).Msgf("/create failed to get queue url for queue name: %s", queueName)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create job of type %s", taskType)})
		return
	}

	_, err = r.sqs.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    queueUrlOutput.QueueUrl,
		MessageBody: &id,
	})
	if err != nil {
		log.Err(err).Msgf("/create failed to get send message to queue url: %s", *queueUrlOutput.QueueUrl)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create job of type %s", taskType)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": id})
}

func (r *Router) status(ctx *gin.Context) {
	id := ctx.Query("id")
	if len(id) < 1 {
		log.Error().Msg("/status query paramter 'id' not passed or empty")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'id' required"})
		return
	}

	status, err := r.das.GetJobStatusByJobId(id)
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

func (r *Router) results(ctx *gin.Context) {
	id := ctx.Param("id")
	if len(id) < 1 {
		log.Error().Msg("/results query paramter 'id' not passed or empty")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'id' required"})
		return
	}

	// TODO
	// 1. verify Job exists and has complete status in DB
	// 2. stream result file from s3
}
