package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Handler) healthzHandler(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "application/text", []byte("ok\n"))
}

func (a *Handler) readyzHandler(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "application/text", []byte("ok\n"))
}
