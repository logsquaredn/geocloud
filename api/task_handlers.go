package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
	"github.com/rs/zerolog/log"
)

func (a *API) listTasksHandler(ctx *gin.Context) {
	tasks, err := a.ds.GetTasks(
		geocloud.AllTaskTypes...,
	)
	if err != nil {
		log.Err(err).Msg("unable to get tasks")
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, tasks)
}

func (a *API) getTaskHandler(ctx *gin.Context) {
	task, statusCode, err := a.getTask(ctx, ctx.Param("type"))
	if err != nil {
		log.Err(err).Msg("unable to get task")
		ctx.AbortWithStatus(statusCode)
		return
	}

	ctx.JSON(http.StatusOK, task)
}
