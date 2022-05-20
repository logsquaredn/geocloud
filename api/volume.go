package api

import (
	"fmt"
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
		inputName       = js.Ternary(applicationZip, "input.zip", "input.json")
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
		contentType     string
		b               []byte
	)
	// if the request was for both content types,
	// we don't know what to do
	if applicationJSON && applicationZip {
		return nil, "", http.StatusBadRequest, fmt.Errorf("only one Content-Type among 'application/json', 'application/zip' may be set")
	}

	err := volume.Walk(func(_ string, f geocloud.File, e error) error {
		switch {
		// pass errors through
		case e != nil:
		// 1. if the request was for application/zip and we found a zip, give it
		// 2. if the request was for application/json and we found some json, give it
		// 3. if the request was for no specific Content-Type, give whatever we found. May be overridden later in the Walk by a zip
		// 4. if the request was for no specific Content-Type and we found a zip, give it
		case applicationZip && filepath.Ext(f.Name()) == ".zip", applicationJSON && filepath.Ext(f.Name()) == ".json", !applicationJSON && !applicationZip && len(b) == 0, !applicationJSON && !applicationZip && filepath.Ext(f.Name()) == ".zip":
			// TODO use io.ReadAll
			// see github.com/logsquaredn/geocloud/objectstore/s3_volume.go
			b = make([]byte, f.Size())
			_, e = f.Read(b)
			contentType = js.Ternary(filepath.Ext(f.Name()) == ".zip", "application/zip", "application/json")
		}
		return e
	})
	switch {
	case err != nil:
		return nil, "", http.StatusInternalServerError, err
	case len(b) == 0:
		return nil, "", http.StatusNotFound, fmt.Errorf("could not find: '%s' content", contentType)
	}

	return b, contentType, 0, nil
}
