package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller/pb"
)

func (a *Handler) checkStorageOwnership(storage *pb.Storage, namespace string) (*pb.Storage, error) {
	if storage.Namespace != namespace {
		return nil, pb.NewErr(fmt.Errorf("requester does not own storage '%s'", storage.Id), http.StatusForbidden)
	}

	return storage, nil
}

func (a *Handler) getStorageForNamespace(id string, namespace string) (*pb.Storage, error) {
	storage, err := a.Datastore.GetStorage(id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, pb.NewErr(fmt.Errorf("storage '%s' not found", id), http.StatusNotFound)
	case err != nil:
		return nil, err
	}

	return a.checkStorageOwnership(storage, namespace)
}

func (a *Handler) createStorageForNamespace(name string, namespace string) (*pb.Storage, error) {
	storage, err := a.Datastore.CreateStorage(&pb.Storage{
		Namespace: namespace,
		Name:      name,
	})
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (a *Handler) getJobOutputStorage(ctx *gin.Context, id string) (*pb.Storage, error) {
	namespace, err := a.getNamespaceFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return a.getJobOutputStorageForNamespace(ctx, id, namespace)
}

func (a *Handler) getJobOutputStorageForNamespace(ctx *gin.Context, id string, namespace string) (*pb.Storage, error) {
	storage, err := a.Datastore.GetJobOutputStorage(id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, pb.NewErr(fmt.Errorf("storage '%s' not found", id), http.StatusNotFound)
	case err != nil:
		return nil, err
	}

	return a.checkStorageOwnership(storage, namespace)
}

func (a *Handler) getJobInputStorage(ctx *gin.Context, id string) (*pb.Storage, error) {
	namespace, err := a.getNamespaceFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return a.getJobInputStorageForNamespace(ctx, id, namespace)
}

func (a *Handler) getJobInputStorageForNamespace(ctx *gin.Context, id string, namespace string) (*pb.Storage, error) {
	storage, err := a.Datastore.GetJobInputStorage(id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, pb.NewErr(fmt.Errorf("storage '%s' not found", id), http.StatusNotFound)
	case err != nil:
		return nil, err
	}

	return a.checkStorageOwnership(storage, namespace)
}
