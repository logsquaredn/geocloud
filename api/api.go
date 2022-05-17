package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud/datastore"
	"github.com/logsquaredn/geocloud/messagequeue"
	"github.com/logsquaredn/geocloud/objectstore"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

type API struct {
	ds     *datastore.Postgres
	mq     *messagequeue.AMQP
	os     *objectstore.S3
	router *gin.Engine
}

func NewServer(opts *Opts) (*API, error) {
	var (
		a = &API{
			ds:     opts.Datastore,
			os:     opts.Objectstore,
			mq:     opts.MessageQueue,
			router: gin.Default(),
		}
	)

	swagger := a.router.Group("/swagger")
	{
		swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	api := a.router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			task := v1.Group("/task")
			{
				task.GET("/", a.listTasksHandler)
				task.GET("/:type", a.getTaskHandler)
			}
			authenticated := v1.Group("/")
			{
				authenticated.Use(a.customerMiddleware)
				storage := authenticated.Group("/storage")
				{
					storage.POST("/", a.createStorageHandler)
					storage.GET("/", a.listStorageHandler)
					storage.GET("/:id", a.getStorageHandler)
					storage.GET("/:id/content", a.getStorageContentHandler)
				}
				job := authenticated.Group("/job")
				{
					job.POST("/buffer", a.createBufferJobHandler)
					job.POST("/filter", a.createFilterJobHandler)
					job.POST("/reproject", a.createReprojectJobHandler)
					job.POST("/removebadgeometry", a.createRemoveBadGeometryJobHandler)
					job.POST("/vectorlookup", a.createVectorLookupJobHandler)
					job.GET("/", a.listJobHandler)
					job.GET("/:id", a.getJobHandler)
					job.GET("/:id/input", a.getJobInputHandler)
					job.GET("/:id/output", a.getJobOutputHandler)
					job.GET("/:id/input/content", a.getJobInputContentHandler)
					job.GET("/:id/output/content", a.getJobOutputContentHandler)
					job.GET("/:id/task", a.getJobTaskHandler)
				}
			}
		}
	}

	return a, nil
}

func (a *API) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	a.router.ServeHTTP(w, req)
}
