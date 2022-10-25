package service

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/pkg/api"
)

// @Security  ApiKeyAuth
// @Summary   Get a list of task types
// @Tags      Task
// @Produce   application/json
// @Success   200  {object}  []rototiller.Task
// @Failure   401  {object}  rototiller.Error
// @Failure   500  {object}  rototiller.Error
// @Router    /api/v1/tasks [get].
func (a *Handler) listTasksHandler(ctx *gin.Context) {
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

// @Security  ApiKeyAuth
// @Summary   Get a task type
// @Tags      Task
// @Produce   application/json
// @Param     type  path      string  true  "Task type"
// @Success   200   {object}  rototiller.Task
// @Failure   400   {object}  rototiller.Error
// @Failure   401   {object}  rototiller.Error
// @Failure   404   {object}  rototiller.Error
// @Failure   500   {object}  rototiller.Error
// @Router    /api/v1/tasks/{type} [get].
func (a *Handler) getTaskHandler(ctx *gin.Context) {
	task, err := a.getTask(ctx.Param("task"))
	if err != nil {
		a.err(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, task)
}
