package service

import (
	"errors"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller/pkg/api"
)

var (
	APIKeyHeader    = "X-API-Key" //nolint:gosec // not an API key, just the header to check for API keys
	errUnauthorized = api.NewErr(
		errors.New(APIKeyHeader+" header must be a valid API key"),
		http.StatusUnauthorized, int(connect.CodeUnauthenticated),
	)
)

// getCustomerFromAPIKey given an API key, actually checks the database for the customer.
func (a *Handler) getCustomerFromAPIKey(apiKey string) (*api.Customer, error) {
	return a.Datastore.GetCustomer(apiKey)
}

// getCustomerFromGinContext given an Gin context, actually checks the database for the customer.
func (a *Handler) getCustomerFromGinContext(ctx *gin.Context) (*api.Customer, error) {
	c, err := a.getCustomerFromAPIKey(a.getCustomerIDFromContext(ctx))
	if err != nil {
		return nil, errUnauthorized
	}

	return c, nil
}

// getAPIKeyFromHeader gets the API key from the given http.Header.
func (a *Handler) getAPIKeyFromHeader(header http.Header) string {
	return header.Get(APIKeyHeader)
}

// getCustomerFromHeader given an http.Header, actually checks the database for the customer.
func (a *Handler) getCustomerFromHeader(header http.Header) (*api.Customer, error) {
	c, err := a.getCustomerFromAPIKey(a.getAPIKeyFromHeader(header))
	if err != nil {
		return nil, errUnauthorized
	}

	return c, nil
}

// getAssumedCustomer returns a customer not hydrated from the database,
// making the assumption that previous middleware has already confirmed
// the customer's existence.
func (a *Handler) getAssumedCustomerFromHeader(header http.Header) *api.Customer {
	return &api.Customer{Id: a.getAPIKeyFromHeader(header)}
}

// getAssumedCustomerFromContext returns a customer not hydrated from the database
// by extracting the customer ID from gin's context.
func (a *Handler) getAssumedCustomerFromContext(ctx *gin.Context) *api.Customer {
	return a.getAssumedCustomerFromHeader(ctx.Request.Header)
}

// customerMiddleare checks for the customer's existence in the database
// and returns a 401 if not found.
func (a *Handler) customerMiddleware(ctx *gin.Context) {
	if _, err := a.getCustomerFromGinContext(ctx); err != nil {
		a.err(ctx, api.NewErr(err, http.StatusUnauthorized, int(connect.CodeUnauthenticated)))
		ctx.Abort()
		return
	}

	ctx.Next()
}

// getCustomerIDFromContext is a duplicate method that
// returns the customer's API key which is also the customer ID.
// The duplication is meant to illustrate this point.
func (a *Handler) getCustomerIDFromContext(ctx *gin.Context) string {
	return a.getAPIKeyFromContext(ctx)
}

// getAPIKeyFromContext gets a customer's API key from gin's context.
func (a *Handler) getAPIKeyFromContext(ctx *gin.Context) string {
	return a.getAPIKeyFromHeader(ctx.Request.Header)
}
