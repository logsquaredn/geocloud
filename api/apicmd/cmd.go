package apicmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/logsquaredn/geocloud/shared/das"
	"github.com/rs/zerolog/log"
)

type Postgres struct {
	Host     string `long:"host" description:"Postgres host"`
	Port     int    `long:"port" default:"5432" description:"Postgres port"`
	User     string `long:"user" default:"geocloud" description:"Postgres username"`
	Password string `long:"password" description:"Postgres password"`
}

type APICmd struct {
	Version func() `short:"v" long:"version" description:"Print the version"`

	Postgres `group:"postgres" namespace:"postgres"`
}

type ConnectionHandler struct {
	das *das.Das
}

func (cmd *APICmd) getDBConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d?sslmode=disable", cmd.Postgres.User, cmd.Postgres.Password, cmd.Postgres.Host, cmd.Postgres.Port)
}

func isJSON(jsBytes []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(jsBytes, &js) == nil
}

func (connHandler *ConnectionHandler) createJob(jobType string) (id string, err error) {
	id = uuid.New().String()
	return id, connHandler.das.InsertNewJob(id, jobType)
}

func (connHandler *ConnectionHandler) buffer(ctx *gin.Context) {
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
	id, err := connHandler.createJob(jobType)
	if err != nil {
		log.Err(err).Msgf("/buffer failed to create job with id: %s and type: %s", id, jobType)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create %s job", jobType)})
		return
	}

	// TODO write jsonData to s3

	// TODO send SQS message

	ctx.JSON(http.StatusOK, gin.H{"id": id})
}

func (connHandler *ConnectionHandler) status(ctx *gin.Context) {
	id := ctx.Query("id")
	if len(id) < 1 {
		log.Error().Msg("/status query paramter 'id' not passed or empty")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'id' required"})
		return
	}
	var status string

	status, err := connHandler.das.GetJobStatusByJobId(id)
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

func (cmd *APICmd) Execute(args []string) error {
	das, err := das.New(das.WithConnectionString(cmd.getDBConnectionString()))
	if err != nil {
		log.Err(err).Msg("failed to connect to DB")
		return err
	}
	defer das.Close()

	connHandler := &ConnectionHandler{
		das: das,
	}

	// TODO format gin middleware logs equivalently to our zerolog format

	router := gin.Default()

	v1_job := router.Group("/api/v1/job")
	{
		v1_job.POST("/buffer", connHandler.buffer)
		v1_job.GET("/status", connHandler.status)
	}

	return router.Run()
}
