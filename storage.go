package geocloud

import (
	"fmt"
	"path"
	"strings"
	"time"
)

type StorageStatus string

const (
	StorageStatusFinal         StorageStatus = "final"
	StorageStatusUnknown       StorageStatus = "unknown"
	StorageStatusUnusable      StorageStatus = "unusable"
	StorageStatusTransformable StorageStatus = "transformable"
)

func (k StorageStatus) String() string {
	return string(k)
}

func ParseStorageStatus(storageStatus string) (StorageStatus, error) {
	for _, k := range []StorageStatus{
		StorageStatusFinal, StorageStatusUnknown,
		StorageStatusUnusable, StorageStatusTransformable,
	} {
		if strings.EqualFold(storageStatus, k.String()) {
			return k, nil
		}
	}

	return "", fmt.Errorf("unknown storage status '%s'", storageStatus)
}

type Storage struct {
	ID         string        `json:"id,omitempty"`
	Status     StorageStatus `json:"kind,omitempty"`
	CustomerID string        `json:"-"`
	Name       string        `json:"name,omitempty"`
	LastUsed   time.Time     `json:"last_used,omitempty"`
	CreateTime time.Time     `json:"create_time,omitempty"`
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

func (c *Client) CreateStorage(r Request) (*Storage, error) {
	var (
		url     = c.baseURL
		storage = &Storage{}
	)

	url.Path = path.Join(EndpointStorage)
	values := url.Query()
	for k, v := range r.Query() {
		if k != "" && v != "" {
			values.Add(k, v)
		}
	}
	url.RawQuery = values.Encode()

	return storage, c.post(url, r, r.ContentType(), storage)
}
