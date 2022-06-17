package api

import (
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud/api/err/v1"
	"github.com/rs/zerolog/log"
)

func (a *API) err(ctx *gin.Context, err error) {
	e := errv1.New(err)
	log.Err(e).Msgf("returning HTTP '%d': %s", e.HTTPStatusCode, e.Message)
	ctx.JSON(e.HTTPStatusCode, e)
}
