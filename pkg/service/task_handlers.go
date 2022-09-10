package service

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller/pkg/api"
)

// @Summary      Get a list of task types
// @Description  &emsp; - API Key is required either as a query parameter or a header
// @Tags         Task
// @Produce      application/json
// @Success      200  {object}  []rototiller.Task
// @Failure      401  {object}  api.Error
// @Failure      500  {object}  api.Error
// @Router       /api/v1/tasks [get].
func (a *API) listTasksHandler(ctx *gin.Context) {
	tasks, err := a.Datastore.GetTasks(
		api.AllTaskTypes...,
	)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		tasks = []*api.Task{}
	case err != nil:
		a.err(ctx, err)
		return
	case tasks == nil:
		tasks = []*api.Task{}
	}

	ctx.JSON(http.StatusOK, tasks)
}

// @Summary      Get a task type
// @Description  &emsp; - API Key is required either as a query parameter or a header
// @Tags         Task
// @Produce      application/json
// @Param        type  path      string  true  "Task type"
// @Success      200   {object}  rototiller.Task
// @Failure      400   {object}  api.Error
// @Failure      401   {object}  api.Error
// @Failure      404   {object}  api.Error
// @Failure      500   {object}  api.Error
// @Router       /api/v1/tasks/{type} [get].
func (a *API) getTaskHandler(ctx *gin.Context) {
	task, err := a.getTask(ctx.Param("task"))
	if err != nil {
		a.err(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, task)
}
