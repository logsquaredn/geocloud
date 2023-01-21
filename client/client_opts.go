package client

import (
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	bufferSize   int
	url          *url.URL
	httpClient   *http.Client
	pollInterval time.Duration
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
