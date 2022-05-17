package api

import (
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
	"github.com/rs/zerolog/log"
)

func (a *API) err(ctx *gin.Context, statusCode int, err error) {
	log.Err(err).Msgf("returning HTTP '%d'", statusCode)
	ctx.AbortWithStatusJSON(statusCode, &geocloud.Error{
		Message: err.Error(),
	})
}
