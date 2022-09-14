package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller/pkg/api"
)

type createCustomerQuery struct {
	email string `form:"email" binding:"required"`
}

// @Summary  Create a new customer
// @Tags     Customer
// @Accept   application/json
// @Produce  application/json
// @Param    email  query     string  true  "Customer email"
// @Success  200    {object}  rototiller.Customer
// @Failure  500    {object}  api.Error
// @Router   /api/v1/customers/create [get].
func (a *Handler) createCustomerHandler(ctx *gin.Context) {
	if err := ctx.ShouldBindQuery(&createCustomerQuery{}); err != nil {
		a.err(ctx, api.NewErr(err, http.StatusBadRequest))
		return
	}

	email := ctx.Query("email")
	if email == "" {
		a.err(ctx, api.NewErr(fmt.Errorf("email cannot be empty"), http.StatusBadRequest))
	}

	c, err := a.createCustomer(email)
	if err != nil {
		a.err(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, c)
}
