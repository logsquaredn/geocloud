package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
)

func (a *API) listTasksHandler(ctx *gin.Context) {
	tasks, err := a.ds.GetTasks(
		geocloud.AllTaskTypes...,
	)
	if err != nil {
		a.err(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, tasks)
}

func (a *API) getTaskHandler(ctx *gin.Context) {
	task, statusCode, err := a.getTask(ctx, ctx.Param("type"))
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	ctx.JSON(http.StatusOK, task)
}
