package geocloud

import (
	"path"
	"time"
)

type Storage struct {
	ID         string    `json:"id,omitempty"`
	CustomerID string    `json:"-"`
	Name       string    `json:"name,omitempty"`
	LastUsed   time.Time `json:"last_used,omitempty"`
}

var _ Message = (*Storage)(nil)

func (s *Storage) GetID() string {
	return s.ID
}

func (c *Client) GetStorages() ([]*Storage, error) {
	var (
		url     = c.baseURL
		storage = []*Storage{}
	)

	url.Path = EndpointStorage

	return storage, c.get(url, &storage)
}

func (c *Client) GetStorage(id string) (*Storage, error) {
	var (
		url     = c.baseURL
		storage = &Storage{}
	)

	url.Path = path.Join(EndpointStorage, id)

	return storage, c.get(url, storage)
}

func (c *Client) GetJobInput(id string) (*Storage, error) {
	var (
		url     = c.baseURL
		storage = &Storage{}
	)

	url.Path = path.Join(EndpointJob, id, "input")

	return storage, c.get(url, storage)
}

func (c *Client) GetJobOutput(id string) (*Storage, error) {
	var (
		url     = c.baseURL
		storage = &Storage{}
	)

	url.Path = path.Join(EndpointJob, id, "output")

	return storage, c.get(url, storage)
}

func (c *Client) CreateStorage(b []byte) (*Storage, error) {
	var (
		url     = c.baseURL
		storage = &Storage{}
	)

	url.Path = path.Join(EndpointStorage)

	return storage, c.post(url, b, storage)
}
