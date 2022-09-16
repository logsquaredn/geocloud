package service

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/gin-gonic/gin"
	_ "github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/internal/rpcio"
	"github.com/logsquaredn/rototiller/pkg/api"
)

// @Security     ApiKeyAuth
// @Summary      Get a list of storage
// @Description  Get a list of stored datasets based on API Key
// @Tags         Storage
// @Produce      application/json
// @Param        offset  query     int  false  "Offset of storages to return"
// @Param        limit   query     int  false  "Limit of storages to return"
// @Success      200     {object}  []rototiller.Storage
// @Failure      401     {object}  rototiller.Error
// @Failure      500     {object}  rototiller.Error
// @Router       /api/v1/storages [get].
func (a *Handler) listStorageHandler(ctx *gin.Context) {
	q := &listQuery{}
	if err := ctx.BindQuery(q); err != nil {
		a.err(ctx, err)
		return
	}

	ownerID, err := a.getOwnerIDFromContext(ctx)
	if err != nil {
		a.err(ctx, err)
		return
	}
	storage, err := a.Datastore.GetOwnerStorage(ownerID, q.Offset, q.Limit)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		storage = []*api.Storage{}
	case err != nil:
		a.err(ctx, err)
		return
	case storage == nil:
		storage = []*api.Storage{}
	}

	ctx.JSON(http.StatusOK, storage)
}

// @Security     ApiKeyAuth
// @Summary      Get a storage
// @Description  Get the metadata of a stored dataset
// @Tags         Storage
// @Produce      application/json
// @Param        id   path      string  true  "Storage ID"
// @Success      200  {object}  rototiller.Storage
// @Failure      401  {object}  rototiller.Error
// @Failure      403  {object}  rototiller.Error
// @Failure      404  {object}  rototiller.Error
// @Failure      500  {object}  rototiller.Error
// @Router       /api/v1/storages/{id} [get].
func (a *Handler) getStorageHandler(ctx *gin.Context) {
	ownerID, err := a.getOwnerIDFromContext(ctx)
	if err != nil {
		a.err(ctx, err)
		return
	}
	storage, err := a.getStorageForOwner(ctx.Param("storage"), ownerID)
	if err != nil {
		a.err(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, storage)
}

// @Security     ApiKeyAuth
// @Summary      Get a storage's content
// @Description  Gets the content of a stored dataset
// @Tags         Content
// @Produce      application/json, application/zip
// @Param        Accept  header  string  false  "Request results as a Zip or JSON. Default Zip"
// @Param        id      path    string  true   "Storage ID"
// @Success      200
// @Failure      400  {object}  rototiller.Error
// @Failure      401  {object}  rototiller.Error
// @Failure      403  {object}  rototiller.Error
// @Failure      404  {object}  rototiller.Error
// @Failure      500  {object}  rototiller.Error
// @Router       /api/v1/storages/{id}/content [get].
func (a *Handler) getStorageContentHandler(ctx *gin.Context) {
	ownerID, err := a.getOwnerIDFromContext(ctx)
	if err != nil {
		a.err(ctx, err)
		return
	}
	storage, err := a.getStorageForOwner(ctx.Param("storage"), ownerID)
	if err != nil {
		a.err(ctx, err)
		return
	}

	volume, err := a.Blobstore.GetObject(ctx, storage.GetId())
	if err != nil {
		a.err(ctx, err)
		return
	}

	r, contentType, err := a.getVolumeContent(ctx.GetHeader("Accept"), volume)
	if err != nil {
		a.err(ctx, err)
		return
	}
	defer r.Close()

	ctx.Writer.Header().Add("Content-Type", contentType)
	_, _ = io.Copy(ctx.Writer, r)
}

// @Security     ApiKeyAuth
// @Summary      Create a storage
// @Description  Stores a dataset. The ID of this stored dataset can be used as input to jobs
// @Description  &emsp; - Pass the geospatial data to be stored in the request body
// @Tags         Storage
// @Accept       application/json, application/zip
// @Produce      application/json
// @Param        name  query     string  false  "Storage name"
// @Success      200   {object}  rototiller.Storage
// @Failure      400   {object}  rototiller.Error
// @Failure      401   {object}  rototiller.Error
// @Failure      500   {object}  rototiller.Error
// @Router       /api/v1/storages [post].
func (a *Handler) createStorageHandler(ctx *gin.Context) {
	defer ctx.Request.Body.Close()
	volume, err := a.getRequestVolume(ctx.Request.Header.Get("Content-Type"), ctx.Request.Body)
	if err != nil {
		a.err(ctx, err)
		return
	}

	ownerID, err := a.getOwnerIDFromContext(ctx)
	if err != nil {
		a.err(ctx, err)
		return
	}
	storage, err := a.createStorageForOwner(ctx.Query("name"), ownerID)
	if err != nil {
		a.err(ctx, err)
		return
	}

	if err = a.Blobstore.PutObject(ctx, storage.GetId(), volume); err != nil {
		a.err(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, storage)
}

// @Security     ApiKeyAuth
// @Summary      RPC Get a storage's content
// @Description  RPC Gets the content of a stored dataset
// @Tags         Content
// @Produce      application/json, application/zip
// @Param        Accept  header    string  false  "Request results as a Zip or JSON. Default Zip"
// @Failure      2       {object}  rototiller.Error
// @Failure      16      {object}  rototiller.Error
// @Router       /api.storage.v1.StorageService/GetStorageContent [post].
func (a *Handler) GetStorageContent(ctx context.Context, req *connect.Request[api.GetStorageContentRequest], stream *connect.ServerStream[api.GetStorageContentResponse]) error {
	ownerID, err := a.getOwnerIDFromHeader(req.Header())
	if err != nil {
		return api.NewConnectErr(err)
	}

	storage, err := a.getStorageForOwner(req.Msg.Id, ownerID)
	if err != nil {
		return api.NewConnectErr(err)
	}

	volume, err := a.Blobstore.GetObject(ctx, storage.GetId())
	if err != nil {
		return api.NewConnectErr(err)
	}

	r, contentType, err := a.getVolumeContent(req.Header().Get("Accept"), volume)
	if err != nil {
		return api.NewConnectErr(err)
	}
	defer r.Close()

	stream.ResponseHeader().Set("X-Content-Type", contentType)
	_, err = io.Copy(
		rpcio.NewServerStreamWriter(
			stream,
			func(b []byte) *api.GetStorageContentResponse {
				return &api.GetStorageContentResponse{
					Data: b,
				}
			},
		),
		r,
	)
	if err != nil {
		return api.NewConnectErr(err)
	}

	return nil
}

// @Security     ApiKeyAuth
// @Summary      RPC Get a storage
// @Description  RPC Get the metadata of a stored dataset
// @Tags         Storage
// @Produce      application/json
// @Failure      2   {object}  rototiller.Error
// @Failure      5   {object}  rototiller.Error
// @Failure      16  {object}  rototiller.Error
// @Router       /api.storage.v1.StorageService/GetStorage [post].
func (a *Handler) GetStorage(ctx context.Context, req *connect.Request[api.GetStorageRequest]) (*connect.Response[api.GetStorageResponse], error) {
	ownerID, err := a.getOwnerIDFromHeader(req.Header())
	if err != nil {
		return nil, api.NewConnectErr(err)
	}

	storage, err := a.getStorageForOwner(req.Msg.GetId(), ownerID)
	if err != nil {
		return nil, api.NewConnectErr(err)
	}

	return connect.NewResponse(&api.GetStorageResponse{
		Storage: &api.Storage{
			Id:   storage.Id,
			Name: storage.Name,
		},
	}), nil
}

// @Security     ApiKeyAuth
// @Summary      RPC Create a storage
// @Description  RPC Stores a dataset. The ID of this stored dataset can be used as input to jobs
// @Tags         Storage
// @Accept       application/json, application/zip
// @Produce      application/json
// @Param        X-Content-Type  header    string  true  "Content type to be stored"
// @Failure      2               {object}  rototiller.Error
// @Failure      5               {object}  rototiller.Error
// @Failure      16              {object}  rototiller.Error
// @Router       /api.storage.v1.StorageService/CreateStorage [post].
func (a *Handler) CreateStorage(ctx context.Context, stream *connect.ClientStream[api.CreateStorageRequest]) (*connect.Response[api.CreateStorageResponse], error) {
	ownerID, err := a.getOwnerIDFromHeader(stream.RequestHeader())
	if err != nil {
		return nil, api.NewConnectErr(err)
	}

	volume, err := a.getRequestVolume(
		stream.RequestHeader().Get("X-Content-Type"),
		rpcio.NewClientStreamReader(
			stream,
			func(t *api.CreateStorageRequest) []byte {
				return t.GetData()
			},
		),
	)
	if err != nil {
		return nil, api.NewConnectErr(err)
	}

	storage, err := a.createStorageForOwner(
		stream.RequestHeader().Get("X-Storage-Name"),
		ownerID,
	)
	if err != nil {
		return nil, api.NewConnectErr(err)
	}

	if err = a.Blobstore.PutObject(ctx, storage.GetId(), volume); err != nil {
		return nil, api.NewConnectErr(err)
	}

	return connect.NewResponse(&api.CreateStorageResponse{
		Storage: &api.Storage{
			Id:   storage.Id,
			Name: storage.Name,
		},
	}), nil
}
