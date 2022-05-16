package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
)

// @Summary Get a list of storage
// @Description
// @Tags
// @Produce application/json
// @Success 200 {object} []geocloud.Storage
// @Failure 401 {object} geocloud.Error
// @Failure 500 {object} geocloud.Error
// @Router /storage [get]
func (a *API) listStorageHandler(ctx *gin.Context) {
	storage, err := a.ds.GetCustomerStorage(a.getAssumedCustomer(ctx))
	if err != nil {
		a.err(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, storage)
}

// @Summary Get a storage
// @Description
// @Tags
// @Produce application/json
// @Param id path string true "Storage ID"
// @Success 200 {object} geocloud.Storage
// @Failure 401 {object} geocloud.Error
// @Failure 403 {object} geocloud.Error
// @Failure 404 {object} geocloud.Error
// @Failure 500 {object} geocloud.Error
// @Router /storage/{id} [get]
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

// @Summary Create a storage
// @Description
// @Tags
// @Accept application/json, application/zip
// @Produce application/json
// @Param name query string false "Storage name"
// @Success 200 {object} geocloud.Storage
// @Failure 400 {object} geocloud.Error
// @Failure 401 {object} geocloud.Error
// @Failure 500 {object} geocloud.Error
// @Router /storage [post]
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

// @Summary Get a storage's content
// @Description
// @Tags
// @Produce application/json
// @Param id path string true "Storage ID"
// @Success 200 {object} geocloud.Job
// @Failure 400 {object} geocloud.Error
// @Failure 401 {object} geocloud.Error
// @Failure 403 {object} geocloud.Error
// @Failure 404 {object} geocloud.Error
// @Failure 500 {object} geocloud.Error
// @Router /storage/{id}/content [get]
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
