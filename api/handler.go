package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/store/blob/bucket"
	"github.com/logsquaredn/rototiller/store/data/postgres"
	"github.com/logsquaredn/rototiller/stream/event/amqp"
	files "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	Datastore           *postgres.Datastore
	EventStreamProducer *amqp.EventStreamProducer
	Blobstore           *bucket.Blobstore
	*http.ServeMux
}

func NewHandler(ctx context.Context, datastore *postgres.Datastore, eventStreamProducer *amqp.EventStreamProducer, blobstore *bucket.Blobstore) (*Handler, error) {
	var (
		logger = rototiller.LoggerFrom(ctx)
		a      = &Handler{
			Datastore:           datastore,
			EventStreamProducer: eventStreamProducer,
			Blobstore:           blobstore,
			ServeMux:            http.NewServeMux(),
		}
		router = gin.New()
	)

	router.Use(gin.Recovery())

	router.GET("/healthz", a.healthzHandler)
	router.GET("/readyz", a.readyzHandler)

	swaggerHandler := swagger.WrapHandler(files.Handler)

	swagger := router.Group("/swagger")
	{
		v1 := swagger.Group("/v1")
		{
			v1.GET("/*any", func(ctx *gin.Context) {
				if ctx.Param("any") == "/" || ctx.Param("any") == "" {
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
				jobs.GET("", a.listJobHandler)
				jobs.POST("/buffer", a.createBufferJobHandler)
				jobs.POST("/filter", a.createFilterJobHandler)
				jobs.POST("/reproject", a.createReprojectJobHandler)
				jobs.POST("/removebadgeometry", a.createRemoveBadGeometryJobHandler)
				jobs.POST("/vectorlookup", a.createVectorLookupJobHandler)
				jobs.POST("/rasterlookup", a.createRasterLookupJobHandler)
				jobs.POST("/polygonvectorlookup", a.createPolygonVectorLookupJobHandler)
				job := jobs.Group("/:job")
				{
					job.GET("", a.getJobHandler)
					job.GET("/tasks", a.getJobTasksHandler)
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

	for _, f := range []func(*Handler) (string, http.Handler){
		func(a *Handler) (string, http.Handler) {
			return "/", router
		},
	} {
		path, handler := f(a)
		a.ServeMux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r.WithContext(rototiller.WithLogger(r.Context(), logger.WithValues("requestId", uuid.NewString()))))
		}))
	}

	return a, nil
}
