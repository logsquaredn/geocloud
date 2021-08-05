package router

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
	"github.com/rs/zerolog/log"
)

func (r *Router) createJob(taskType string) (geocloud.Job, error) {
	return r.das.InsertJob(taskType)
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
	task, err := r.das.GetTaskByTaskType(taskType)
	if err == sql.ErrNoRows {
		log.Error().Msgf("/create invalid task type requested: %s", taskType)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid task type: %s", taskType)})
		return
	} else if err != nil {
		log.Err(err).Msgf("/create failed to query for params for type: %s", taskType)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create job of type %s", taskType)})
		return
	}

	missingParams := validateParamsPassed(ctx, task.Params)
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

	job, err := r.createJob(taskType)
	if err != nil {
		log.Err(err).Msgf("/create failed to create job of type: %s", taskType)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create job of type %s", taskType)})
		return
	}

	_, err = r.oas.PutJobInput(job.ID, bytes.NewReader(jsonData), "geojson")
	if err != nil {
		log.Err(err).Msgf("/create failed to write data to s3 for id: %s", job.ID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create job of type %s", taskType)})
		return
	}

	queueUrlOutput, err := r.sqs.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &task.QueueName,
	})
	if err != nil {
		log.Err(err).Msgf("/create failed to get queue url for queue name: %s", task.QueueName)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create job of type %s", taskType)})
		return
	}

	_, err = r.sqs.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    queueUrlOutput.QueueUrl,
		MessageBody: &job.ID,
	})
	if err != nil {
		log.Err(err).Msgf("/create failed to get send message to queue url: %s", *queueUrlOutput.QueueUrl)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create job of type %s", taskType)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": job.ID})
}

func (r *Router) status(ctx *gin.Context) {
	id := ctx.Query("id")
	if len(id) < 1 {
		log.Error().Msg("/status query paramter 'id' not passed or empty")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'id' required"})
		return
	}

	job, err := r.das.GetJobByJobID(id)
	if err == sql.ErrNoRows {
		log.Err(err).Msgf("/status got 0 results querying for id: %s", id)
		ctx.Status(http.StatusNotFound)
		return
	} else if err != nil {
		log.Err(err).Msgf("/status failed to query for status for id: %s", id)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get status for id: %s", id)})
		return
	}

	responseBody := gin.H{"status": job.Status}
	if job.Status == geocloud.Error {
		responseBody["error"] = job.Error.Error()
	}
	ctx.JSON(http.StatusOK, responseBody)
}

func (r *Router) result(ctx *gin.Context) {
	id := ctx.Query("id")
	if len(id) < 1 {
		log.Error().Msg("/result query paramter 'id' not passed or empty")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'id' required"})
		return
	}

	job, err := r.das.GetJobByJobID(id)
	if err == sql.ErrNoRows {
		log.Err(err).Msgf("/result got 0 results querying for id: %s", id)
		ctx.Status(http.StatusNotFound)
		return
	} else if err != nil {
		log.Err(err).Msgf("/result failed to query for status for id: %s", id)
		ctx.Status(http.StatusInternalServerError)
		return
	} else if job.Status != geocloud.Completed {
		log.Error().Msgf("/result results requested but id: %s is of status: %s", id, job.Status)
		ctx.Status(http.StatusBadRequest)
		return
	}

	buf := aws.NewWriteAtBuffer([]byte{})
	err = r.oas.GetJobOutput(id, buf, "geojson")
	if err != nil {
		log.Error().Msgf("/result failed to download result from s3 for id: %s", id)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	var js map[string]interface{}
	json.Unmarshal(buf.Bytes(), &js)
	if js == nil {
		log.Error().Msgf("/result failed to convert result to valid json for id: %s", id)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, js)
}
