package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
)

func (a *API) checkStorageOwnership(ctx *gin.Context, storage *geocloud.Storage) (*geocloud.Storage, int, error) {
	return a.checkStorageOwnershipForCustomer(ctx, storage, a.getAssumedCustomer(ctx))
}

func (a *API) checkStorageOwnershipForCustomer(ctx *gin.Context, storage *geocloud.Storage, customer *geocloud.Customer) (*geocloud.Storage, int, error) {
	if storage.CustomerID != customer.ID {
		// you could make an argument for 404 here
		// as "idk what storage you're talking about"
		// is more secure than "there's definitely a storage here
		// but you can't see it"
		return nil, http.StatusForbidden, fmt.Errorf("customer does not own storage")
	}

	return storage, 0, nil
}

func (a *API) getStorage(ctx *gin.Context, m geocloud.Message) (*geocloud.Storage, int, error) {
	return a.getStorageForCustomer(ctx, m, a.getAssumedCustomer(ctx))
}

func (a *API) getStorageForCustomer(ctx *gin.Context, m geocloud.Message, customer *geocloud.Customer) (*geocloud.Storage, int, error) {
	storage, err := a.ds.GetStorage(m)
	switch {
	case err != nil:
		return nil, http.StatusNotFound, err
	}

	return a.checkStorageOwnershipForCustomer(ctx, storage, customer)
}

func (a *API) createStorage(ctx *gin.Context) (*geocloud.Storage, int, error) {
	return a.createStorageForCustomer(ctx, a.getAssumedCustomer(ctx))
}

func (a *API) createStorageForCustomer(ctx *gin.Context, customer *geocloud.Customer) (*geocloud.Storage, int, error) {
	storage, err := a.ds.CreateStorage(&geocloud.Storage{
		CustomerID: customer.ID,
	})
	if err != nil {
		return nil, 500, err
	}

	return storage, 0, nil
}

func (a *API) getJobOutputStorage(ctx *gin.Context, m geocloud.Message) (*geocloud.Storage, int, error) {
	return a.getJobOutputStorageForCustomer(ctx, m, a.getAssumedCustomer(ctx))
}

func (a *API) getJobOutputStorageForCustomer(ctx *gin.Context, m geocloud.Message, customer *geocloud.Customer) (*geocloud.Storage, int, error) {
	storage, err := a.ds.GetJobOutputStorage(m)
	if err != nil {
		ctx.AbortWithStatus(http.StatusNotFound)
	}

	return a.checkStorageOwnershipForCustomer(ctx, storage, customer)
}

func (a *API) getJobInputStorage(ctx *gin.Context, m geocloud.Message) (*geocloud.Storage, int, error) {
	return a.getJobInputStorageForCustomer(ctx, m, a.getAssumedCustomer(ctx))
}

func (a *API) getJobInputStorageForCustomer(ctx *gin.Context, m geocloud.Message, customer *geocloud.Customer) (*geocloud.Storage, int, error) {
	storage, err := a.ds.GetJobInputStorage(m)
	if err != nil {
		ctx.AbortWithStatus(http.StatusNotFound)
	}

	return a.checkStorageOwnershipForCustomer(ctx, storage, customer)
}
