package service

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller/internal/rpcio"
	"github.com/logsquaredn/rototiller/pkg/api"
)

// @Summary      Get a list of storage
// @Description  Get a list of stored datasets based on API Key
// @Description
// @Description  &emsp; - API Key is required either as a query parameter or a header
// @Tags         Storage
// @Produce      application/json
// @Param        X-API-Key       header    string  false  "API Key header"
// @Param        api-key    query     string  false  "API Key query parameter"
// @Param        offset     query     int     false  "Offset of storages to return"
// @Param        limit      query     int     false  "Limit of storages to return"
// @Success      200        {object}  []rototiller.Storage
// @Failure      401        {object}  api.Error
// @Failure      500        {object}  api.Error
// @Router       /api/v1/storages [get].
func (a *API) listStorageHandler(ctx *gin.Context) {
	q := &listQuery{}
	if err := ctx.BindQuery(q); err != nil {
		a.err(ctx, err)
		return
	}

	storage, err := a.Datastore.GetCustomerStorage(a.getAssumedCustomerFromContext(ctx).GetId(), q.Offset, q.Limit)
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

// @Summary      Get a storage
// @Description  Get the metadata of a stored dataset
// @Description
// @Description  &emsp; - API Key is required either as a query parameter or a header
// @Tags         Storage
// @Produce      application/json
// @Param        api-key    query     string  false  "API Key query parameter"
// @Param        X-API-Key  header    string  false  "API Key header"
// @Param        id         path      string  true   "Storage ID"
// @Success      200        {object}  rototiller.Storage
// @Failure      401        {object}  api.Error
// @Failure      403        {object}  api.Error
// @Failure      404        {object}  api.Error
// @Failure      500        {object}  api.Error
// @Router       /api/v1/storages/{id} [get].
func (a *API) getStorageHandler(ctx *gin.Context) {
	var (
		storage, err = a.getStorageForCustomer(ctx.Param("storage"), a.getAssumedCustomerFromContext(ctx))
	)
	if err != nil {
		a.err(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, storage)
}

// @Summary      Get a storage's content
// @Description  Gets the content of a stored dataset
// @Description
// @Description  &emsp; - API Key is required either as a query parameter or a header
// @Tags         Content
// @Produce      application/json, application/zip
// @Param        Accept     header  string  false  "Request results as a Zip or JSON. Default Zip"
// @Param        api-key    query   string  false  "API Key query parameter"
// @Param        X-API-Key  header  string  false  "API Key header"
// @Param        id         path    string  true   "Storage ID"
// @Success      200
// @Failure      400  {object}  api.Error
// @Failure      401  {object}  api.Error
// @Failure      403  {object}  api.Error
// @Failure      404  {object}  api.Error
// @Failure      500  {object}  api.Error
// @Router       /api/v1/storages/{id}/content [get].
func (a *API) getStorageContentHandler(ctx *gin.Context) {
	storage, err := a.getStorageForCustomer(ctx.Param("storage"), a.getAssumedCustomerFromContext(ctx))
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

// @Summary      Create a storage
// @Description  Stores a dataset. The ID of this stored dataset can be used as input to jobs
// @Description
// @Description  &emsp; - API Key is required either as a query parameter or a header
// @Description  &emsp; - Pass the geospatial data to be stored in the request body
// @Tags         Storage
// @Accept       application/json, application/zip
// @Produce      application/json
// @Param        api-key    query     string  false  "API Key query parameter"
// @Param        X-API-Key  header    string  false  "API Key header"
// @Param        name       query     string  false  "Storage name"
// @Success      200        {object}  rototiller.Storage
// @Failure      400        {object}  api.Error
// @Failure      401        {object}  api.Error
// @Failure      500        {object}  api.Error
// @Router       /api/v1/storages [post].
func (a *API) createStorageHandler(ctx *gin.Context) {
	defer ctx.Request.Body.Close()
	volume, err := a.getRequestVolume(ctx.Request.Header.Get("Content-Type"), ctx.Request.Body)
	if err != nil {
		a.err(ctx, err)
		return
	}

	storage, err := a.createStorageForCustomer(ctx.Query("name"), a.getAssumedCustomerFromContext(ctx))
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

// @Summary      RPC Get a storage's content
// @Description  RPC Gets the content of a stored dataset
// @Tags         Content
// @Produce      application/json, application/zip
// @Param        Accept     header    string  false  "Request results as a Zip or JSON. Default Zip"
// @Param        X-API-Key  header    string  false  "API Key header"
// @Failure      2               {object}  api.Error
// @Failure      16              {object}  api.Error
// @Router       /api.storage.v1.StorageService/GetStorageContent [post].
func (a *API) GetStorageContent(ctx context.Context, req *connect.Request[api.GetStorageContentRequest], stream *connect.ServerStream[api.GetStorageContentResponse]) error {
	// TODO refactor into interceptor https://connect.build/docs/go/streaming#interceptors
	_, err := a.getCustomerFromConnectHeader(req.Header())
	if err != nil {
		return connect.NewError(api.NewErr(err).ConnectCode, err)
	}

	storage, err := a.getStorageForCustomer(req.Msg.Id, a.getAssumedCustomer(req.Header().Get("X-API-Key")))
	if err != nil {
		return connect.NewError(api.NewErr(err).ConnectCode, err)
	}

	volume, err := a.Blobstore.GetObject(ctx, storage.GetId())
	if err != nil {
		return connect.NewError(api.NewErr(err).ConnectCode, err)
	}

	r, contentType, err := a.getVolumeContent(req.Header().Get("Accept"), volume)
	if err != nil {
		return connect.NewError(api.NewErr(err).ConnectCode, err)
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
		return connect.NewError(api.NewErr(err).ConnectCode, err)
	}

	return nil
}

// @Summary      RPC Get a storage
// @Description  RPC Get the metadata of a stored dataset
// @Tags         Storage
// @Produce      application/json
// @Param        X-API-Key  header    string  false  "API Key header"
// @Failure      2          {object}  api.Error
// @Failure      5               {object}  api.Error
// @Failure      16         {object}  api.Error
// @Router       /api.storage.v1.StorageService/GetStorage [post].
func (a *API) GetStorage(ctx context.Context, req *connect.Request[api.GetStorageRequest]) (*connect.Response[api.GetStorageResponse], error) {
	_, err := a.getCustomerFromConnectHeader(req.Header())
	if err != nil {
		return nil, connect.NewError(api.NewErr(err).ConnectCode, err)
	}

	storage, err := a.getStorageForCustomer(req.Msg.GetId(), a.getAssumedCustomer(req.Header().Get("X-API-Key")))
	if err != nil {
		return nil, connect.NewError(api.NewErr(err).ConnectCode, err)
	}

	return connect.NewResponse(&api.GetStorageResponse{
		Storage: &api.Storage{
			Id:   storage.Id,
			Name: storage.Name,
		},
	}), nil
}

// @Summary      RPC Create a storage
// @Description  RPC Stores a dataset. The ID of this stored dataset can be used as input to jobs
// @Tags         Storage
// @Accept       application/json, application/zip
// @Produce      application/json
// @Param        X-Content-Type  header    string  true   "Content type to be stored"
// @Param        X-API-Key  header    string  false  "API Key header"
// @Failure      2          {object}  api.Error
// @Failure      5          {object}  api.Error
// @Failure      16         {object}  api.Error
// @Router       /api.storage.v1.StorageService/CreateStorage [post].
func (a *API) CreateStorage(ctx context.Context, stream *connect.ClientStream[api.CreateStorageRequest]) (*connect.Response[api.CreateStorageResponse], error) {
	_, err := a.getCustomerFromConnectHeader(stream.RequestHeader())
	if err != nil {
		return nil, connect.NewError(api.NewErr(err).ConnectCode, err)
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
		return nil, connect.NewError(api.NewErr(err).ConnectCode, err)
	}

	storage, err := a.createStorageForCustomer(
		stream.RequestHeader().Get("X-Storage-Name"),
		a.getAssumedCustomer(stream.RequestHeader().Get("X-API-Key")),
	)
	if err != nil {
		return nil, connect.NewError(api.NewErr(err).ConnectCode, err)
	}

	if err = a.Blobstore.PutObject(ctx, storage.GetId(), volume); err != nil {
		return nil, connect.NewError(api.NewErr(err).ConnectCode, err)
	}

	return connect.NewResponse(&api.CreateStorageResponse{
		Storage: &api.Storage{
			Id:   storage.Id,
			Name: storage.Name,
		},
	}), nil
}
