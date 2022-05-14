package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud/datastore"
	"github.com/logsquaredn/geocloud/messagequeue"
	"github.com/logsquaredn/geocloud/objectstore"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

type API struct {
	ds     *datastore.Postgres
	mq     *messagequeue.AMQP
	os     *objectstore.S3
	router *gin.Engine
}


func NewServer(opts *APIOpts) (*API, error) {
	var (
		a = &API{
			ds:     opts.Datastore,
			os:     opts.Objectstore,
			mq:     opts.MessageQueue,
			router: gin.Default(),
		}
	)

	a.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	a.router.Use(a.middleware)

	v1Job := a.router.Group("/api/v1/job")
	{
		v1Job.POST("/create/buffer", a.createBuffer)
		v1Job.POST("/create/removebadgeometry", a.createRemovebadgeometry)
		v1Job.POST("/create/reproject", a.createReproject)
		v1Job.POST("/create/filter", a.createFilter)
		v1Job.POST("/create/vectorlookup", a.createVectorlookup)
		v1Job.GET("/status", a.status)
		v1Job.GET("/result", a.result)
	}

	return a, nil
}

func (a *API) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	a.router.ServeHTTP(w, req)
}
