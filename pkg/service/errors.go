package service

import (
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/pkg/api"
)

func (a *Handler) err(ctx *gin.Context, err error) {
	logr := rototiller.LoggerFrom(ctx)

	apiErr := api.NewErr(err)
	logr.Error(apiErr, "")
	ctx.JSON(apiErr.HTTPStatusCode, apiErr)
}
