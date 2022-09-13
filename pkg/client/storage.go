package client

import (
	"context"
	"io"
	"path"

	"github.com/logsquaredn/rototiller/internal/rpcio"
	"github.com/logsquaredn/rototiller/pkg/api"
)

func (c *Client) GetStorages() ([]*api.Storage, error) {
	var (
		url     = c.url
		storage = []*api.Storage{}
	)

	url.Path = api.EndpointStorages

	return storage, c.get(url, &storage)
}

func (c *Client) GetStorage(id string) (*api.Storage, error) {
	var (
		url     = c.url
		storage = &api.Storage{}
	)

	url.Path = path.Join(api.EndpointStorages, id)

	return storage, c.get(url, storage)
}

func (c *Client) GetJobInput(id string) (*api.Storage, error) {
	var (
		url     = c.url
		storage = &api.Storage{}
	)

	url.Path = path.Join(api.EndpointJobs, id, "storages", "input")

	return storage, c.get(url, storage)
}

func (c *Client) GetJobOutput(id string) (*api.Storage, error) {
	var (
		url     = c.url
		storage = &api.Storage{}
	)

	url.Path = path.Join(api.EndpointJobs, id, "storages", "output")

	return storage, c.get(url, storage)
}

func (c *Client) CreateStorage(ctx context.Context, r Request) (*api.Storage, error) {
	if c.rpc {
		var (
			stream = c.storageClient.CreateStorage(ctx)
		)
		stream.RequestHeader().Add("X-Content-Type", r.ContentType())
		stream.RequestHeader().Add("X-Storage-Name", r.Query()["name"])

		if _, err := io.CopyBuffer(
			rpcio.NewClientStreamWriter(stream, func(b []byte) *api.CreateStorageRequest {
				return &api.CreateStorageRequest{
					Data: b,
				}
			}),
			r,
			make([]byte, c.bufferSize),
		); err != nil {
			return nil, err
		}

		res, err := stream.CloseAndReceive()
		if err != nil {
			return nil, err
		}

		return &api.Storage{
			Id:   res.Msg.Storage.Id,
			Name: res.Msg.Storage.Name,
		}, nil
	}

	var (
		url     = c.url
		storage = &api.Storage{}
	)

	url.Path = api.EndpointStorages
	values := url.Query()
	for k, v := range r.Query() {
		if k != "" && v != "" {
			values.Add(k, v)
		}
	}
	url.RawQuery = values.Encode()

	return storage, c.post(url, r, r.ContentType(), storage)
}
