package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
)

// @Summary Get a list of task types
// @Description
// @Tags
// @Produce application/json
// @Success 200 {object} geocloud.Job
// @Failure 401 {object} geocloud.Error
// @Failure 500 {object} geocloud.Error
// @Router /task [get]
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

// @Summary Get a task type
// @Description
// @Tags
// @Produce application/json
// @Param type path string true "Task type"
// @Success 200 {object} geocloud.Job
// @Failure 400 {object} geocloud.Error
// @Failure 401 {object} geocloud.Error
// @Failure 404 {object} geocloud.Error
// @Failure 500 {object} geocloud.Error
// @Router /task/{type} [post]
func (a *API) getTaskHandler(ctx *gin.Context) {
	task, statusCode, err := a.getTask(ctx, ctx.Param("type"))
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	ctx.JSON(http.StatusOK, task)
}
