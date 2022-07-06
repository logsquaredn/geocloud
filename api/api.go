package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller/api/storage/v1/storagev1connect"
	"github.com/logsquaredn/rototiller/datastore"
	"github.com/logsquaredn/rototiller/messagequeue"
	"github.com/logsquaredn/rototiller/objectstore"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

type API struct {
	ds  *datastore.Postgres
	mq  *messagequeue.AMQP
	os  *objectstore.S3
	mux *http.ServeMux

	storagev1connect.UnimplementedStorageServiceHandler
}

func NewServer(opts *Opts) (*API, error) {
	var (
		a = &API{
			ds:  opts.Datastore,
			os:  opts.Objectstore,
			mq:  opts.MessageQueue,
			mux: http.NewServeMux(),
		}
		router = gin.Default()
	)

	router.GET("/healthz", a.healthzHandler)
	router.GET("/readyz", a.readyzHandler)

	swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)

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
			return storagev1connect.NewStorageServiceHandler(a)
		},
		func(a *API) (string, http.Handler) {
			return "/", router
		},
	} {
		path, handler := f(a)
		a.mux.Handle(path, handler)
	}

	return a, nil
}

func (a *API) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	a.mux.ServeHTTP(w, req)
}
