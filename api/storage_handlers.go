package api

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
	storagev1 "github.com/logsquaredn/geocloud/api/storage/v1"
)

// @Summary      Get a list of storage
// @Description  Get a list of stored datasets based on API Key
// @Description
// @Description  &emsp; - API Key is required either as a query parameter or a header
// @Tags         Storage
// @Produce      application/json
// @Param        api-key    query     string  false  "API Key via query parameter"
// @Param        X-API-Key  header    string  false  "API Key via header"
// @Success      200        {object}  []geocloud.Storage
// @Failure      401        {object}  geocloud.Error
// @Failure      500        {object}  geocloud.Error
// @Router       /storage [get]
func (a *API) listStorageHandler(ctx *gin.Context) {
	storage, err := a.ds.GetCustomerStorage(a.getAssumedCustomer(ctx))
	switch {
	case errors.Is(err, sql.ErrNoRows):
		storage = []*geocloud.Storage{}
	case err != nil:
		a.err(ctx, http.StatusInternalServerError, err)
	case storage == nil:
		storage = []*geocloud.Storage{}
	}

	ctx.JSON(http.StatusOK, storage)
}

// @Summary      Get a storage
// @Description  Get the metadata of a stored dataset
// @Description
// @Description  &emsp; - API Key is required either as a query parameter or a header
// @Tags         Storage
// @Produce      application/json
// @Param        api-key    query     string  false  "API Key via query parameter"
// @Param        X-API-Key  header    string  false  "API Key via header"
// @Param        id         path      string  true   "Storage ID"
// @Success      200        {object}  geocloud.Storage
// @Failure      401        {object}  geocloud.Error
// @Failure      403        {object}  geocloud.Error
// @Failure      404        {object}  geocloud.Error
// @Failure      500        {object}  geocloud.Error
// @Router       /storage/{id} [get]
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

// @Summary      Get a storage's content
// @Description  Gets the content of a stored dataset
// @Description
// @Description  &emsp; - API Key is required either as a query parameter or a header
// @Tags         Content
// @Produce      application/json, application/zip
// @Param        Content-Type  header  string  false  "Request results as a Zip or JSON. Default Zip"
// @Param        api-key       query   string  false  "API Key via query parameter"
// @Param        X-API-Key     header  string  false  "API Key via header"
// @Param        id            path    string  true   "Storage ID"
// @Success      200
// @Failure      400  {object}  geocloud.Error
// @Failure      401  {object}  geocloud.Error
// @Failure      403  {object}  geocloud.Error
// @Failure      404  {object}  geocloud.Error
// @Failure      500  {object}  geocloud.Error
// @Router       /storage/{id}/content [get]
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
		a.err(ctx, http.StatusInternalServerError, err)
		return
	}

	r, contentType, statusCode, err := a.getVolumeContent(ctx, volume)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}
	defer r.Close()

	ctx.Writer.Header().Add("Content-Type", contentType)
	_, _ = io.Copy(ctx.Writer, r)
}

func (a *API) GetStorage(ctx context.Context, req *connect.Request[storagev1.GetStorageRequest], stream *connect.ServerStream[storagev1.GetStorageResponse]) error {
	// TODO
	// for i := 0; i < 5; i++ {
	// 	res := &storagev1.GetStorageResponse{}
	// 	res.Data = []byte("blah")

	// 	err := stream.Send(res)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	time.Sleep(3 * time.Second)
	// }

	return nil
}

func (a *API) CreateStorage(ctx context.Context, stream *connect.ClientStream[storagev1.CreateStorageRequest]) (*connect.Response[storagev1.CreateStorageResponse], error) {
	buf := []byte{}
	for stream.Receive() {
		buf = append(buf, stream.Msg().Data...)
	}
	if err := stream.Err(); err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	// TODO wrap the stream in io.Reader interface and pass it directly here
	volume, _, err := a.getRequestVolume(stream.RequestHeader().Get("X-Content-Type"), bytes.NewReader(buf))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	storage, err := a.ds.CreateStorage(&geocloud.Storage{
		CustomerID: stream.RequestHeader().Get("X-API-Key"),
		Name:       stream.RequestHeader().Get("X-Storage-Name"),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	if err = a.os.PutObject(storage, volume); err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	// TODO return actual storage object
	return connect.NewResponse(&storagev1.CreateStorageResponse{
		Id: storage.ID,
	}), nil
}

// @Summary      Create a storage
// @Description  Stores a dataset. The ID of this stored dataset can be used as input to jobs
// @Description
// @Description  &emsp; - API Key is required either as a query parameter or a header
// @Description  &emsp; - Pass the geospatial data to be stored in the request body
// @Tags         Storage
// @Accept       application/json, application/zip
// @Produce      application/json
// @Param        api-key    query     string  false  "API Key via query parameter"
// @Param        X-API-Key  header    string  false  "API Key via header"
// @Param        name       query     string  false  "Storage name"
// @Success      200        {object}  geocloud.Storage
// @Failure      400        {object}  geocloud.Error
// @Failure      401        {object}  geocloud.Error
// @Failure      500        {object}  geocloud.Error
// @Router       /storage [post]
func (a *API) createStorageHandler(ctx *gin.Context) {
	defer ctx.Request.Body.Close()
	volume, statusCode, err := a.getRequestVolume(ctx.Request.Header.Get("Content-Type"), ctx.Request.Body)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	storage, statusCode, err := a.createStorage(ctx)
	if err != nil {
		a.err(ctx, statusCode, err)
		return
	}

	if err = a.os.PutObject(storage, volume); err != nil {
		a.err(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, storage)
}
