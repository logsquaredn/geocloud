package geocloud

import (
	"net/http"
	"net/url"
)

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

type ClientOpt func(*Client) error

func WithHTTPClient(httpClient *http.Client) ClientOpt {
	return func(c *Client) error {
		c.httpClient = httpClient
		return nil
	}
}
