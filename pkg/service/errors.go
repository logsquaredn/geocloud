package service

import (
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller/pkg/api"
)

func (a *Handler) err(ctx *gin.Context, err error) {
	apiErr := api.NewErr(err)
	ctx.JSON(apiErr.HTTPStatusCode, apiErr)
}
