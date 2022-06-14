package api

import (
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
	"github.com/rs/zerolog/log"
)

type apiError struct {
	message        string
	httpStatusCode int
	rpcCode        uint32
}

func (e *apiError) Error() string {
	return e.message
}

func (a *API) err(ctx *gin.Context, statusCode int, err error) {
	log.Err(err).Msgf("returning HTTP '%d'", statusCode)
	ctx.JSON(statusCode, &geocloud.Error{
		Message: err.Error(),
	})
}
