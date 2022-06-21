package api

import (
	"fmt"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
	errv1 "github.com/logsquaredn/geocloud/api/err/v1"
)

// getCustomerFromApiKey given an API key, actually checks the database for the customer
func (a *API) getCustomerFromAPIKey(apiKey string) (*geocloud.Customer, error) {
	c, err := a.ds.GetCustomer(
		geocloud.Msg(apiKey),
	)
	if err != nil {
		return nil, errv1.New(
			fmt.Errorf("header '%s' must be a valid API Key", geocloud.APIKeyHeader),
			http.StatusUnauthorized, int(connect.CodeUnauthenticated),
		)
	}

	return c, nil
}

// getCustomerFromGinContext given an Gin context, actually checks the database for the customer
func (a *API) getCustomerFromGinContext(ctx *gin.Context) (*geocloud.Customer, error) {
	c, err := a.getCustomerFromAPIKey(a.getCustomerIDFromContext(ctx))
	if err != nil {
		return nil, errv1.New(
			fmt.Errorf(
				"query parameter '%s', header '%s' or cookie '%s' must be a valid API Key",
				geocloud.APIKeyQueryParam, geocloud.APIKeyHeader, geocloud.APIKeyCookie,
			),
			http.StatusUnauthorized, int(connect.CodeUnauthenticated),
		)
	}

	return c, nil
}

// getCustomerFromConnectHeader given a Connect header, actually checks the database for the customer
func (a *API) getCustomerFromConnectHeader(header http.Header) (*geocloud.Customer, error) {
	c, err := a.getCustomerFromAPIKey(header.Get(geocloud.APIKeyHeader))
	if err != nil {
		return nil, errv1.New(
			fmt.Errorf("header '%s' must be a valid API Key", geocloud.APIKeyHeader),
			http.StatusUnauthorized, int(connect.CodeUnauthenticated),
		)
	}

	return c, nil
}

// getAssumedCustomer returns a customer not hydrated from the database,
// making the assumption that previous middleware has already confirmed
// the customer's existence
func (a *API) getAssumedCustomer(id string) *geocloud.Customer {
	return &geocloud.Customer{ID: id}
}

// getAssumedCustomerFromContext returns a customer not hydrated from the database
// by extracting the customer ID from gin's context
func (a *API) getAssumedCustomerFromContext(ctx *gin.Context) *geocloud.Customer {
	return a.getAssumedCustomer(a.getCustomerIDFromContext(ctx))
}

// customerMiddleare checks for the customer's existence in the database
// and returns a 401 if not found
func (a *API) customerMiddleware(ctx *gin.Context) {
	if _, err := a.getCustomerFromGinContext(ctx); err != nil {
		a.err(ctx, errv1.New(err, http.StatusUnauthorized, int(connect.CodeUnauthenticated)))
		ctx.Abort()
		return
	}

	ctx.Next()
}

// getCustomerIDFromContext is a duplicate method that
// returns the customer's API key which is also the customer ID.
// The duplication is meant to illustrate this point
func (a *API) getCustomerIDFromContext(ctx *gin.Context) string {
	return a.getAPIKeyFromContext(ctx)
}

// getAPIKeyFromContext gets a customer's API key from gin's context
func (a *API) getAPIKeyFromContext(ctx *gin.Context) string {
	apiKey := ctx.Query(geocloud.APIKeyQueryParam)
	if apiKey == "" {
		apiKey = ctx.GetHeader(geocloud.APIKeyHeader)
		if apiKey == "" {
			apiKey, _ = ctx.Cookie(geocloud.APIKeyCookie)
		}
	}

	return apiKey
}
