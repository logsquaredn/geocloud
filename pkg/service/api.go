package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller/pkg/api/apiconnect"
	"github.com/logsquaredn/rototiller/pkg/store/blob/bucket"
	"github.com/logsquaredn/rototiller/pkg/store/data/postgres"
	"github.com/logsquaredn/rototiller/pkg/stream/event/amqp"
	files "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
)

type API struct {
	Datastore           *postgres.Datastore
	EventStreamProducer *amqp.EventStreamProducer
	Blobstore           *bucket.Blobstore
	*http.ServeMux
	apiconnect.UnimplementedStorageServiceHandler
}

func NewServer(datastore *postgres.Datastore, eventStreamProducer *amqp.EventStreamProducer, blobstore *bucket.Blobstore) (*API, error) {
	var (
		a = &API{
			Datastore:           datastore,
			EventStreamProducer: eventStreamProducer,
			Blobstore:           blobstore,
			ServeMux:            http.NewServeMux(),
		}
		router = gin.Default()
	)

	router.GET("/healthz", a.healthzHandler)
	router.GET("/readyz", a.readyzHandler)

	swaggerHandler := swagger.WrapHandler(files.Handler)

	router.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/swagger/v1/index.html")
	})

	swagger := router.Group("/swagger")
	{
		v1 := swagger.Group("/v1")
		{
			v1.GET("/*any", func(ctx *gin.Context) {
				if ctx.Param("any") == "" {
					ctx.Redirect(http.StatusFound, "/swagger/v1/index.html")
				} else {
					swaggerHandler(ctx)
				}
			})
		}
	}

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			tasks := v1.Group("/tasks")
			{
				tasks.GET("", a.listTasksHandler)
				tasks.GET("/:task", a.getTaskHandler)
			}
			storages := v1.Group("/storages")
			{
				storages.Use(a.customerMiddleware)
				storages.POST("", a.createStorageHandler)
				storages.GET("", a.listStorageHandler)
				storage := storages.Group("/:storage")
				{
					storage.GET("", a.getStorageHandler)
					storage.GET("/content", a.getStorageContentHandler)
				}
			}
			jobs := v1.Group("/jobs")
			{
				jobs.Use(a.customerMiddleware)
				jobs.GET("", a.listJobHandler)
				jobs.POST("/buffer", a.createBufferJobHandler)
				jobs.POST("/filter", a.createFilterJobHandler)
				jobs.POST("/reproject", a.createReprojectJobHandler)
				jobs.POST("/removebadgeometry", a.createRemoveBadGeometryJobHandler)
				jobs.POST("/vectorlookup", a.createVectorLookupJobHandler)
				jobs.POST("/rasterlookup", a.createRasterLookupJobHandler)
				jobs.POST("/polygonVectorLookup", a.createPolygonVectorLookupJobHandler)
				job := jobs.Group("/:job")
				{
					job.GET("", a.getJobHandler)
					job.GET("/task", a.getJobTaskHandler)
					jobStorages := job.Group("storages")
					{
						jobStorageInput := jobStorages.Group("input")
						{
							jobStorageInput.GET("", a.getJobInputHandler)
							jobStorageInput.GET("/content", a.getJobInputContentHandler)
						}
						jobStorageOutput := jobStorages.Group("output")
						{
							jobStorageOutput.GET("", a.getJobOutputHandler)
							jobStorageOutput.GET("/content", a.getJobOutputContentHandler)
						}
					}
				}
			}
		}
	}

	for _, f := range []func(*API) (string, http.Handler){
		func(a *API) (string, http.Handler) {
			return apiconnect.NewStorageServiceHandler(a)
		},
		func(a *API) (string, http.Handler) {
			return "/", router
		},
	} {
		path, handler := f(a)
		a.ServeMux.Handle(path, handler)
	}

	return a, nil
}
