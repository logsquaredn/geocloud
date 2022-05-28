package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
)

// @Summary  Get a list of task types
// @Description  &emsp; - API Key is required either as a query parameter or a header
// @Tags     Task
// @Produce  application/json
// @Success  200  {object}  []geocloud.Task
// @Failure  401  {object}  geocloud.Error
// @Failure  500  {object}  geocloud.Error
// @Router   /task [get]
func (a *API) listTasksHandler(ctx *gin.Context) {
	tasks, err := a.ds.GetTasks(
		geocloud.AllTaskTypes...,
	)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		tasks = []*geocloud.Task{}
	case err != nil:
		a.err(ctx, http.StatusInternalServerError, err)
		return
	case tasks == nil:
		tasks = []*geocloud.Task{}
	}

	ctx.JSON(http.StatusOK, tasks)
}

// @Summary  Get a task type
// @Description  &emsp; - API Key is required either as a query parameter or a header// @Tags     Task
// @Produce  application/json
// @Param    type  path      string  true  "Task type"
// @Success  200   {object}  geocloud.Task
// @Failure  400   {object}  geocloud.Error
// @Failure  401   {object}  geocloud.Error
// @Failure  404   {object}  geocloud.Error
// @Failure  500   {object}  geocloud.Error
// @Router   /task/{type} [get]
func (a *API) getTaskHandler(ctx *gin.Context) {
	task, statusCode, err := a.getTask(ctx, ctx.Param("type"))
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	ctx.JSON(http.StatusOK, task)
}
