package api

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/frantjc/go-js"
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
)

func (a *API) putRequestVolumeForCustomer(ctx *gin.Context, customer *geocloud.Customer) (*geocloud.Storage, int, error) {
	volume, statusCode, err := a.getRequestVolume(ctx)
	if err != nil {
		return nil, statusCode, err
	}

	storage, statusCode, err := a.createStorageForCustomer(ctx, customer)
	if err != nil {
		return nil, statusCode, err
	}

	if err = a.os.PutObject(storage, volume); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return storage, 0, nil
}

func (a *API) getRequestVolume(ctx *gin.Context) (geocloud.Volume, int, error) {
	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	var (
		contentType     = ctx.Request.Header.Get("Content-Type")
		applicationJSON = strings.Contains(contentType, "application/json")
		applicationZip  = strings.Contains(contentType, "application/zip")
		inputName       = js.Ternary(applicationZip, "input.zip", "input.geojson")
	)
	switch {
	case applicationJSON && applicationZip:
		// both possible Content-Types are specified
		return nil, http.StatusBadRequest, err
	case applicationZip && !isZIP(data):
		// Content-Type is application/zip but data is not
		return nil, http.StatusBadRequest, err
	case applicationJSON && !isJSON(data):
		// Content-Type is application/json but data is not
		return nil, http.StatusBadRequest, err
	case !(applicationJSON || applicationZip):
		// neither possible Content-Types are specified
		return nil, http.StatusBadRequest, err
	}

	return geocloud.NewSingleFileVolume(inputName, data), 0, nil
}

func (a *API) getVolumeContent(ctx *gin.Context, volume geocloud.Volume) ([]byte, string, int, error) {
	var (
		accept          = ctx.Request.Header.Get("Content-Type")
		applicationJSON = strings.Contains(accept, "application/json")
		applicationZip  = strings.Contains(accept, "application/zip")
		contentType     = js.Ternary(applicationZip, "application/zip", "application/json")
		b               []byte
	)
	err := volume.Walk(func(_ string, f geocloud.File, e error) error {
		switch {
		case e != nil:
			return e
		case applicationZip && filepath.Ext(f.Name()) == ".zip", applicationJSON && filepath.Ext(f.Name()) == ".geojson":
			// TODO use io.ReadAll
			// see github.com/logsquaredn/geocloud/objectstore/s3_volume.go
			b = make([]byte, f.Size())
			_, e = f.Read(b)
		}
		return e
	})
	switch {
	case err != nil:
		return nil, "", http.StatusInternalServerError, err
	case len(b) == 0:
		return nil, "", http.StatusInternalServerError, err
	}

	return b, contentType, 0, nil
}
