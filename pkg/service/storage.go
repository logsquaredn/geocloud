package service

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller/pkg/api"
)

func (a *Handler) checkStorageOwnership(storage *api.Storage, ownerID string) (*api.Storage, error) {
	if storage.OwnerId != ownerID {
		return nil, api.NewErr(fmt.Errorf("requester does not own storage '%s'", storage.Id), http.StatusForbidden)
	}

	return storage, nil
}

func (a *Handler) getStorageForOwner(id string, ownerID string) (*api.Storage, error) {
	storage, err := a.Datastore.GetStorage(id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, api.NewErr(fmt.Errorf("storage '%s' not found", id), http.StatusNotFound, int(connect.CodeNotFound))
	case err != nil:
		return nil, err
	}

	return a.checkStorageOwnership(storage, ownerID)
}

func (a *Handler) createStorageForOwner(name string, ownerID string) (*api.Storage, error) {
	storage, err := a.Datastore.CreateStorage(&api.Storage{
		OwnerId: ownerID,
		Name:    name,
	})
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (a *Handler) getJobOutputStorage(ctx *gin.Context, id string) (*api.Storage, error) {
	ownerID, err := a.getOwnerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return a.getJobOutputStorageForOwner(ctx, id, ownerID)
}

func (a *Handler) getJobOutputStorageForOwner(ctx *gin.Context, id string, ownerID string) (*api.Storage, error) {
	storage, err := a.Datastore.GetJobOutputStorage(id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, api.NewErr(fmt.Errorf("storage '%s' not found", id), http.StatusNotFound)
	case err != nil:
		return nil, err
	}

	return a.checkStorageOwnership(storage, ownerID)
}

func (a *Handler) getJobInputStorage(ctx *gin.Context, id string) (*api.Storage, error) {
	ownerID, err := a.getOwnerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return a.getJobInputStorageForOwner(ctx, id, ownerID)
}

func (a *Handler) getJobInputStorageForOwner(ctx *gin.Context, id string, ownerID string) (*api.Storage, error) {
	storage, err := a.Datastore.GetJobInputStorage(id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, api.NewErr(fmt.Errorf("storage '%s' not found", id), http.StatusNotFound)
	case err != nil:
		return nil, err
	}

	return a.checkStorageOwnership(storage, ownerID)
}
