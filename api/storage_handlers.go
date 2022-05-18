package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
)

// @Summary  Get a list of storage
// @Description Get a list of stored datasets based on API Key
// @Tags listStorage
// @Produce  application/json
// @Param api-key query string false "API Key via query parameter"
// @Param X-API-Key header string false "API Key via header"
// @Success  200  {object}  []geocloud.Storage
// @Failure  401  {object}  geocloud.Error
// @Failure  500  {object}  geocloud.Error
// @Router   /storage [get]
func (a *API) listStorageHandler(ctx *gin.Context) {
	storage, err := a.ds.GetCustomerStorage(a.getAssumedCustomer(ctx))
	switch {
	case errors.Is(err, sql.ErrNoRows):
		storage = []*geocloud.Storage{}
	case err != nil:
		a.err(ctx, http.StatusInternalServerError, err)
	case storage == nil:
		storage = []*geocloud.Storage{}
	}

	ctx.JSON(http.StatusOK, storage)
}

// @Summary  Get a storage
// @Description Get the metadata of a stored dataset
// @Tags getStorage
// @Produce  application/json
// @Param api-key query string false "API Key via query parameter"
// @Param X-API-Key header string false "API Key via header"
// @Param    id   path      string  true  "Storage ID"
// @Success  200  {object}  geocloud.Storage
// @Failure  401  {object}  geocloud.Error
// @Failure  403  {object}  geocloud.Error
// @Failure  404  {object}  geocloud.Error
// @Failure  500  {object}  geocloud.Error
// @Router   /storage/{id} [get]
func (a *API) getStorageHandler(ctx *gin.Context) {
	var (
		storage, statusCode, err = a.getStorage(
			ctx,
			geocloud.NewMessage(
				ctx.Param("id"),
			),
		)
	)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	ctx.JSON(http.StatusOK, storage)
}

// @Summary  Get a storage's content
// @Description Gets the content of a stored dataset
// @Tags getStorageContent
// @Produce  application/json, application/zip
// @Param Content-Type header string false "Request results as a Zip or JSON. Default Zip"
// @Param api-key query string false "API Key via query parameter"
// @Param X-API-Key header string false "API Key via header"
// @Param    id  path  string  true  "Storage ID"
// @Success  200
// @Failure  400  {object}  geocloud.Error
// @Failure  401  {object}  geocloud.Error
// @Failure  403  {object}  geocloud.Error
// @Failure  404  {object}  geocloud.Error
// @Failure  500  {object}  geocloud.Error
// @Router   /storage/{id}/content [get]
func (a *API) getStorageContentHandler(ctx *gin.Context) {
	storage, statusCode, err := a.getStorage(
		ctx,
		geocloud.NewMessage(
			ctx.Param("id"),
		),
	)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	volume, err := a.os.GetObject(storage)
	if err != nil {
		a.err(ctx, http.StatusInternalServerError, err)
		return
	}

	b, contentType, statusCode, err := a.getVolumeContent(ctx, volume)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	ctx.Data(http.StatusOK, contentType, b)
}

// @Summary Create a storage
// @Description Stores a dataset. The ID of this stored dataset can be used as input to jobs
// @Tags createStorage
// @Accept application/json, application/zip
// @Produce application/json
// @Param api-key query string false "API Key via query parameter"
// @Param X-API-Key header string false "API Key via header"
// @Param    name  query     string  false  "Storage name"
// @Success  200   {object}  geocloud.Storage
// @Failure  400   {object}  geocloud.Error
// @Failure  401   {object}  geocloud.Error
// @Failure  500   {object}  geocloud.Error
// @Router   /storage [post]
func (a *API) createStorageHandler(ctx *gin.Context) {
	storage, statusCode, err := a.createStorage(ctx)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	volume, statusCode, err := a.getRequestVolume(ctx)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	if err = a.os.PutObject(storage, volume); err != nil {
		a.err(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, storage)
}
