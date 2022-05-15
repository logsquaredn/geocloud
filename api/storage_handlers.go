package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
	"github.com/rs/zerolog/log"
)

func (a *API) listStorageHandler(ctx *gin.Context) {
	customer := a.getAssumedCustomer(ctx)
	storage, err := a.ds.GetCustomerStorage(customer)
	if err != nil {
		log.Err(err).Msgf("unable to find storage for customer '%s'", customer.ID)
		ctx.AbortWithStatus(http.StatusInternalServerError)
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
		log.Err(err).Msgf("unable to get storage '%s'", storage.ID)
		ctx.AbortWithStatus(statusCode)
		return
	}

	ctx.JSON(http.StatusOK, storage)
}

func (a *API) createStorageHandler(ctx *gin.Context) {
	storage, statusCode, err := a.createStorage(ctx)
	if err != nil {
		log.Err(err).Msg("unable to create storage")
		ctx.AbortWithStatus(statusCode)
		return
	}

	volume, statusCode, err := a.getRequestVolume(ctx)
	if err != nil {
		log.Err(err).Msg("unable to get request volume")
		ctx.AbortWithStatus(statusCode)
		return
	}

	if err = a.os.PutObject(storage, volume); err != nil {
		log.Err(err).Msg("unable to put volume")
		ctx.AbortWithStatus(http.StatusInternalServerError)
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
		log.Err(err).Msg("unable to get storage")
		ctx.AbortWithStatus(statusCode)
		return
	}

	volume, err := a.os.GetObject(storage)
	if err != nil {
		log.Err(err).Msg("unable to get object")
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	b, contentType, statusCode, err := a.getVolumeContent(ctx, volume)
	if err != nil {
		log.Err(err).Msg("unable to get volume content")
		ctx.AbortWithStatus(statusCode)
		return
	}

	ctx.Data(http.StatusOK, contentType, b)
}
