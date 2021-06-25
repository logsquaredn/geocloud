package apicmd

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type APICmd struct { // variable that ends up getting populated from arguments from the command line
	Postgres struct {
		Password string `long:"password" description:"postgres password"`
		Username string `long:"username" description:"postgres username"`
		Port     int    `long:"port" description:"postgres port"`
		Host     string `long:"host" description:"postgres host"`
	} `group:"postgres" namespace:"postgres"`	
}

type Handler struct {
	conn driver.Conn
}

func (cmd *APICmd) getDBPath() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d", cmd.Postgres.Username, cmd.Postgres.Password, cmd.Postgres.Host, cmd.Postgres.Port)
}
func isJSON(jsBytes []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(jsBytes, &js) == nil
}

func (h *Handler) buffer(ctx *gin.Context) { // gin.context -> contains information from the request we recieve.
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

/*
pass us a job id, use the job id to get information from the database.
1. get the job id from the query parameters and validate that it's not empty.
	- if empty, return a 400.
*/
func (h *Handler) status(ctx *gin.Context) {
	id := ctx.Query("id")
	if len(id) < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'id' required"})
	}
	
	stmt, err := h.conn.Prepare("SELECT job_name FROM job WHERE job_id = 1")
	if err != nil {
		// send error idk how yet
	}

	fmt.Print(stmt)
}
func (cmd *APICmd) Execute(args []string) error {
	conn, err := pq.Open(cmd.getDBPath())
	if err != nil {
		return err
	}

	h := &Handler {
		conn: conn,
	}

	router := gin.Default() // creating an instance of gin web framework. router = manages paths, routes a path to the request it's requested to.

	v1_job := router.Group("/api/v1/job") // everything contained in this group will be under the specified path in the api.
	{
		v1_job.POST("/buffer", h.buffer) // first endpoint, will buffer the endpoint it's given. "/api/v1/job/buffer"
		v1_job.GET("/status", h.status)
	}

	return router.Run() // start's the web application and will run infinitely until it's killed.
}
