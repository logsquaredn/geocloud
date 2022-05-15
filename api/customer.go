package api

import (
	"net/http"

	"github.com/frantjc/go-js"
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
	"github.com/rs/zerolog/log"
)

// getCustomer actually checks the database for the customer
func (a *API) getCustomer(ctx *gin.Context) (*geocloud.Customer, int, error) {
	c, err := a.ds.GetCustomer(
		geocloud.NewMessage(
			getCustomerID(ctx),
		),
	)
	return c, js.Ternary(err == nil, 0, http.StatusUnauthorized), err
}

// getAssumedCustomer returns a customer not hydrated from the database,
// making the assumption that previous middleware has already confirmed
// the customer's existence
func (a *API) getAssumedCustomer(ctx *gin.Context) *geocloud.Customer {
	return &geocloud.Customer{
		ID: getCustomerID(ctx),
	}
}

// customerMiddleare checks for the customer's existence in the database
// and returns a 401 if not found
func (a *API) customerMiddleware(ctx *gin.Context) {
	if _, statusCode, err := a.getCustomer(ctx); err != nil {
		log.Err(err).Msgf("query parameter '%s', header '%s' or cookie '%s' must be a valid API Key", apiKeyQueryParam, apiKeyHeader, apiKeyCookie)
		ctx.AbortWithStatus(statusCode)
		return
	}

	ctx.Next()
}
