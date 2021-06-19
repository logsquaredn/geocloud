package apicmd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

type APICmd struct{}

func isJSON(jsBytes []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(jsBytes, &js) == nil
}

func buffer(context *gin.Context) {
	distance := context.Query("distance")
	if len(distance) < 1 {
		context.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'distance' required"})
	}

	jsonData, err := ioutil.ReadAll(context.Request.Body)
	if err != nil || !isJSON(jsonData) {
		context.JSON(http.StatusBadRequest, gin.H{"error": "request body must be valid json"})
	}

	// TODO create job in DB

	// TODO write jsonData to s3

	// TODO send SQS message

}

func (cmd *APICmd) Execute(args []string) error {
	router := gin.Default()

	v1_job := router.Group("/api/v1/job")
	{
		v1_job.POST("/buffer", buffer)
	}

	return router.Run()
}
