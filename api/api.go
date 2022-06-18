package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud/api/storage/v1/storagev1connect"
	"github.com/logsquaredn/geocloud/datastore"
	"github.com/logsquaredn/geocloud/messagequeue"
	"github.com/logsquaredn/geocloud/objectstore"
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
			task := v1.Group("/task")
			{
				task.GET("", a.listTasksHandler)
				task.GET("/:type", a.getTaskHandler)
			}
			authenticated := v1.Group("/")
			{
				authenticated.Use(a.customerMiddleware)
				storage := authenticated.Group("/storage")
				{
					storage.POST("", a.createStorageHandler)
					storage.GET("", a.listStorageHandler)
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
					job.POST("/rasterlookup", a.createRasterLookupJobHandler)
					job.GET("", a.listJobHandler)
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
