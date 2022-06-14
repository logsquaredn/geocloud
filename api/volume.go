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
	var (
		contentType     = ctx.Request.Header.Get("Content-Type")
		applicationJSON = strings.Contains(contentType, "application/json")
		applicationZip  = strings.Contains(contentType, "application/zip")
		inputName       = js.Ternary(applicationZip, "input.zip", "input.json")
	)
	switch {
	case applicationJSON && applicationZip:
		// both possible Content-Types are specified
		return nil, http.StatusBadRequest, fmt.Errorf("only one Content-Type among 'application/json' and 'application/zip' may be specified")
	case !(applicationJSON || applicationZip):
		// neither possible Content-Types are specified
		return nil, http.StatusBadRequest, fmt.Errorf("must specify one Content-Type among 'application/json' and 'application/zip'")
	}

	return geocloud.NewSingleFileVolume(inputName, ctx.Request.Body), 0, nil
}

func (a *API) getVolumeContent(ctx *gin.Context, volume geocloud.Volume) (io.ReadCloser, string, int, error) {
	var (
		accept          = ctx.Request.Header.Get("Accept")
		applicationJSON = strings.Contains(accept, "application/json")
		applicationZip  = strings.Contains(accept, "application/zip")
		contentType     string
		r               io.ReadCloser
	)
	// if the request was for both content types,
	// we don't know what to do
	if applicationJSON && applicationZip {
		return nil, "", http.StatusBadRequest, fmt.Errorf("only one Accept among 'application/json' and 'application/zip' may be specified")
	}

	err := volume.Walk(func(_ string, f geocloud.File, e error) error {
		switch {
		// pass errors through
		case e != nil:
		// 1. if the request was for application/zip and we found a zip, give it
		// 2. if the request was for application/json and we found some json, give it
		// 3. if the request was for no specific Content-Type, give whatever we found. May be overridden later in the Walk by a zip
		// 4. if the request was for no specific Content-Type and we found a zip, give it
		case applicationZip && filepath.Ext(f.Name()) == ".zip", applicationJSON && filepath.Ext(f.Name()) == ".json", !applicationJSON && !applicationZip && r == nil, !applicationJSON && !applicationZip && filepath.Ext(f.Name()) == ".zip":
			r = f
			contentType = js.Ternary(filepath.Ext(f.Name()) == ".zip", "application/zip", "application/json")
		}
		return e
	})
	switch {
	case err != nil:
		return nil, "", http.StatusInternalServerError, err
	case r == nil:
		return nil, "", http.StatusNotFound, fmt.Errorf("could not find: '%s' content", contentType)
	}

	return r, contentType, 0, nil
}
