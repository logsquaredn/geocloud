package apicmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
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
	db *sql.DB
}

func (cmd *APICmd) getDBConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d?sslmode=disable", cmd.Postgres.User, cmd.Postgres.Password, cmd.Postgres.Host, cmd.Postgres.Port)
}

func isJSON(jsBytes []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(jsBytes, &js) == nil
}

func (connHandler *ConnectionHandler) buffer(ctx *gin.Context) {
	distance := ctx.Query("distance")
	if len(distance) < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'distance' required"})
	}

	jsonData, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil || !isJSON(jsonData) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "request body must be valid json"})
	}

	// TODO create job in DB

	// TODO write jsonData to s3

	// TODO send SQS message

}

func (connHandler *ConnectionHandler) status(ctx *gin.Context) {
	id := ctx.Query("id")
	if len(id) < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'id' required"})
	}
	var status string

	row := connHandler.db.QueryRow("SELECT job_status FROM job WHERE job_id = $1", id)
	err := row.Scan(&status)

	if err == sql.ErrNoRows {
		ctx.Status(http.StatusNotFound)
		return
	} else if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": status})

}
func (cmd *APICmd) Execute(args []string) error {
	db, err := sql.Open("postgres", cmd.getDBConnectionString())
	if err != nil {
		return err
	}

	connHandler := &ConnectionHandler{
		db: db,
	}

	router := gin.Default()

	v1_job := router.Group("/api/v1/job")
	{
		v1_job.POST("/buffer", connHandler.buffer)
		v1_job.GET("/status", connHandler.status)
	}

	return router.Run()
}
