package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
)

func (a *API) listStorageHandler(ctx *gin.Context) {
	storage, err := a.ds.GetCustomerStorage(a.getAssumedCustomer(ctx))
	if err != nil {
		a.err(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, storage)
}

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
		a.err(ctx, statusCode, err)
		return
	}

	ctx.JSON(http.StatusOK, storage)
}

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
		a.err(ctx, statusCode, err)
		return
	}

	b, contentType, statusCode, err := a.getVolumeContent(ctx, volume)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	ctx.Data(http.StatusOK, contentType, b)
}
