package geocloud

import (
	"net/http"
	"net/url"
	"time"

	"github.com/logsquaredn/geocloud/api/storage/v1/storagev1connect"
)

type Client struct {
	rpc           bool
	bufferSize    int
	baseURL       *url.URL
	httpClient    *http.Client
	storageClient storagev1connect.StorageServiceClient
	pollInterval  time.Duration
}

type ClientOpt func(*Client) error

func WithHTTPClient(httpClient *http.Client) ClientOpt {
	return func(c *Client) error {
		c.httpClient = httpClient
		return nil
	}
}

func WithPollInterval(pollInterval time.Duration) ClientOpt {
	return func(c *Client) error {
		c.pollInterval = pollInterval
		return nil
	}
}

func WithBufferSize(bufferSize int) ClientOpt {
	return func(c *Client) error {
		c.bufferSize = bufferSize
		return nil
	}
}

func WithRPC(c *Client) error {
	c.rpc = true
	return nil
}
