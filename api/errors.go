package api

import (
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller/pb"
)

func (a *Handler) err(ctx *gin.Context, err error) {
	apiErr := pb.NewErr(err)
	ctx.JSON(apiErr.HTTPStatusCode, apiErr)
}
