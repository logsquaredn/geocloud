package apicmd

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type APICmd struct{}

func buffer(context *gin.Context) {
	distance := context.Query("distance")

	fmt.Printf("distance: %s", distance)
}

func (cmd *APICmd) Execute(args []string) error {
	router := gin.Default()

	v1_job := router.Group("/v1/job")
	{
		v1_job.POST("/buffer", buffer)
	}

	router.Run()

	return fmt.Errorf("api wip")
}
