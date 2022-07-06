package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller"
	errv1 "github.com/logsquaredn/rototiller/api/err/v1"
)

func (a *API) checkStorageOwnershipForCustomer(storage *rototiller.Storage, customer *rototiller.Customer) (*rototiller.Storage, error) {
	if storage.CustomerID != customer.ID {
		return nil, errv1.New(fmt.Errorf("customer does not own storage '%s'", storage.ID), http.StatusForbidden)
	}

	return storage, nil
}

func (a *API) getStorageForCustomer(m rototiller.Message, customer *rototiller.Customer) (*rototiller.Storage, error) {
	storage, err := a.ds.GetStorage(m)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, errv1.New(fmt.Errorf("storage '%s' not found", m.GetID()), http.StatusNotFound, int(connect.CodeNotFound))
	case err != nil:
		return nil, err
	}

	return a.checkStorageOwnershipForCustomer(storage, customer)
}

func (a *API) createStorageForCustomer(name string, customer *rototiller.Customer) (*rototiller.Storage, error) {
	storage, err := a.ds.CreateStorage(&rototiller.Storage{
		CustomerID: customer.ID,
		Name:       name,
	})
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (a *API) getJobOutputStorage(ctx *gin.Context, m rototiller.Message) (*rototiller.Storage, error) {
	return a.getJobOutputStorageForCustomer(ctx, m, a.getAssumedCustomerFromContext(ctx))
}

func (a *API) getJobOutputStorageForCustomer(ctx *gin.Context, m rototiller.Message, customer *rototiller.Customer) (*rototiller.Storage, error) {
	storage, err := a.ds.GetJobOutputStorage(m)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, errv1.New(fmt.Errorf("storage '%s' not found", m.GetID()), http.StatusNotFound)
	case err != nil:
		return nil, err
	}

	return a.checkStorageOwnershipForCustomer(storage, customer)
}

func (a *API) getJobInputStorage(ctx *gin.Context, m rototiller.Message) (*rototiller.Storage, error) {
	return a.getJobInputStorageForCustomer(ctx, m, a.getAssumedCustomerFromContext(ctx))
}

func (a *API) getJobInputStorageForCustomer(ctx *gin.Context, m rototiller.Message, customer *rototiller.Customer) (*rototiller.Storage, error) {
	storage, err := a.ds.GetJobInputStorage(m)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, errv1.New(fmt.Errorf("storage '%s' not found", m.GetID()), http.StatusNotFound)
	case err != nil:
		return nil, err
	}

	return a.checkStorageOwnershipForCustomer(storage, customer)
}
