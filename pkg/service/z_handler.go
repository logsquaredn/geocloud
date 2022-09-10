package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *API) healthzHandler(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "application/text", []byte("ok\n"))
}

func (a *API) readyzHandler(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "application/text", []byte("ok\n"))
}
