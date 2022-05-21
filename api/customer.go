package api

import (
	"fmt"
	"net/http"

	"github.com/frantjc/go-js"
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
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
		a.err(ctx, statusCode, fmt.Errorf("query parameter '%s', header '%s' or cookie '%s' must be a valid API Key", apiKeyQueryParam, apiKeyHeader, apiKeyCookie))
		ctx.Abort()
		return
	}

	ctx.Next()
}

const (
	apiKeyQueryParam = "api-key"
	apiKeyHeader     = "X-API-Key" //nolint:gosec
	apiKeyCookie     = apiKeyHeader
)

var getCustomerID = getAPIKey

func getAPIKey(ctx *gin.Context) string {
	apiKey := ctx.Query(apiKeyQueryParam)
	if apiKey == "" {
		apiKey = ctx.GetHeader(apiKeyHeader)
		if apiKey == "" {
			apiKey, _ = ctx.Cookie(apiKeyCookie)
		}
	}
	return apiKey
}
