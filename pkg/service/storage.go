package service

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller/pkg/api"
)

func (a *Handler) checkStorageOwnershipForCustomer(storage *api.Storage, customer *api.Customer) (*api.Storage, error) {
	if storage.CustomerId != customer.Id {
		return nil, api.NewErr(fmt.Errorf("customer does not own storage '%s'", storage.Id), http.StatusForbidden)
	}

	return storage, nil
}

func (a *Handler) getStorageForCustomer(id string, customer *api.Customer) (*api.Storage, error) {
	storage, err := a.Datastore.GetStorage(id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, api.NewErr(fmt.Errorf("storage '%s' not found", id), http.StatusNotFound, int(connect.CodeNotFound))
	case err != nil:
		return nil, err
	}

	return a.checkStorageOwnershipForCustomer(storage, customer)
}

func (a *Handler) createStorageForCustomer(name string, customer *api.Customer) (*api.Storage, error) {
	storage, err := a.Datastore.CreateStorage(&api.Storage{
		CustomerId: customer.Id,
		Name:       name,
	})
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (a *Handler) getJobOutputStorage(ctx *gin.Context, id string) (*api.Storage, error) {
	return a.getJobOutputStorageForCustomer(ctx, id, a.getAssumedCustomerFromContext(ctx))
}

func (a *Handler) getJobOutputStorageForCustomer(ctx *gin.Context, id string, customer *api.Customer) (*api.Storage, error) {
	storage, err := a.Datastore.GetJobOutputStorage(id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, api.NewErr(fmt.Errorf("storage '%s' not found", id), http.StatusNotFound)
	case err != nil:
		return nil, err
	}

	return a.checkStorageOwnershipForCustomer(storage, customer)
}

func (a *Handler) getJobInputStorage(ctx *gin.Context, id string) (*api.Storage, error) {
	return a.getJobInputStorageForCustomer(ctx, id, a.getAssumedCustomerFromContext(ctx))
}

func (a *Handler) getJobInputStorageForCustomer(ctx *gin.Context, id string, customer *api.Customer) (*api.Storage, error) {
	storage, err := a.Datastore.GetJobInputStorage(id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, api.NewErr(fmt.Errorf("storage '%s' not found", id), http.StatusNotFound)
	case err != nil:
		return nil, err
	}

	return a.checkStorageOwnershipForCustomer(storage, customer)
}
