package api

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/frantjc/go-js"
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller/pb"
	"github.com/logsquaredn/rototiller/volume"
)

const (
	zipExt  = ".zip"
	jsonExt = ".json"
)

func (a *Handler) putRequestVolumeForNamespace(ctx *gin.Context, contentType, name string, r io.Reader, namespace string) (*pb.Storage, error) {
	volume, err := a.getRequestVolume(contentType, r)
	if err != nil {
		return nil, err
	}

	storage, err := a.createStorageForNamespace(name, namespace)
	if err != nil {
		return nil, err
	}

	if err = a.Blobstore.PutObject(ctx, storage.GetId(), volume); err != nil {
		return nil, err
	}

	return storage, nil
}

func (a *Handler) getRequestVolume(contentType string, r io.Reader) (volume.Volume, error) {
	var (
		applicationJSON = strings.Contains(contentType, "application/json")
		applicationZip  = strings.Contains(contentType, "application/zip")
		inputName       = js.Ternary(applicationZip, "input.zip", "input.json")
	)
	switch {
	case applicationJSON && applicationZip:
		// both possible Content-Types are specified
		return nil, pb.NewErr(fmt.Errorf("only one Content-Type among 'application/json' and 'application/zip' may be specified"), http.StatusBadRequest)
	case !(applicationJSON || applicationZip):
		// neither possible Content-Types are specified
		return nil, pb.NewErr(fmt.Errorf("must specify one Content-Type among 'application/json' and 'application/zip'"), http.StatusBadRequest)
	}

	return volume.New(volume.NewFile(inputName, r, 0)), nil
}

func (a *Handler) getVolumeContent(accept string, vol volume.Volume) (io.ReadCloser, string, error) {
	var (
		applicationJSON = strings.Contains(accept, "application/json")
		applicationZip  = strings.Contains(accept, "application/zip")
		contentType     string
		r               io.ReadCloser
	)
	// if the request was for both content types,
	// we don't know what to do
	if applicationJSON && applicationZip {
		return nil, "", pb.NewErr(fmt.Errorf("only one Accept among 'application/json' and 'application/zip' may be specified"), http.StatusBadRequest)
	}

	err := vol.Walk(func(_ string, f volume.File, e error) error {
		switch {
		// pass errors through
		case e != nil:
		// 1. if the request was for application/zip and we found a zip, give it
		// 2. if the request was for application/json and we found some json, give it
		// 3. if the request was for no specific Content-Type, give whatever we found. May be overridden later in the Walk by a zip
		// 4. if the request was for no specific Content-Type and we found a zip, give it
		case applicationZip && filepath.Ext(f.GetName()) == zipExt, applicationJSON && filepath.Ext(f.GetName()) == jsonExt, !applicationJSON && !applicationZip && r == nil, !applicationJSON && !applicationZip && filepath.Ext(f.GetName()) == zipExt:
			r = f
			contentType = js.Ternary(filepath.Ext(f.GetName()) == zipExt, "application/zip", "application/json")
		}
		return e
	})
	switch {
	case err != nil:
		return nil, "", err
	case r == nil:
		return nil, "", pb.NewErr(fmt.Errorf("could not find: '%s' content", contentType), http.StatusNotFound)
	}

	return r, contentType, nil
}
