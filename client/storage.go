package client

import (
	"context"
	"path"

	"github.com/logsquaredn/rototiller/pb"
)

func (c *Client) GetStorages() ([]*pb.Storage, error) {
	var (
		url     = c.url
		storage = []*pb.Storage{}
	)

	url.Path = pb.EndpointStorages

	return storage, c.get(url, &storage)
}

func (c *Client) GetStorage(id string) (*pb.Storage, error) {
	var (
		url     = c.url
		storage = &pb.Storage{}
	)

	url.Path = path.Join(pb.EndpointStorages, id)

	return storage, c.get(url, storage)
}

func (c *Client) GetJobInput(id string) (*pb.Storage, error) {
	var (
		url     = c.url
		storage = &pb.Storage{}
	)

	url.Path = path.Join(pb.EndpointJobs, id, "storages", "input")

	return storage, c.get(url, storage)
}

func (c *Client) GetJobOutput(id string) (*pb.Storage, error) {
	var (
		url     = c.url
		storage = &pb.Storage{}
	)

	url.Path = path.Join(pb.EndpointJobs, id, "storages", "output")

	return storage, c.get(url, storage)
}

func (c *Client) CreateStorage(ctx context.Context, r Request) (*pb.Storage, error) {
	var (
		url     = c.url
		storage = &pb.Storage{}
	)

	url.Path = pb.EndpointStorages
	values := url.Query()
	for k, v := range r.Query() {
		if k != "" && v != "" {
			values.Add(k, v)
		}
	}
	url.RawQuery = values.Encode()

	return storage, c.post(url, r, r.ContentType(), storage)
}
