package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/user"
	"time"

	"github.com/logsquaredn/rototiller/pb"
)

func New(addr, apiKey string, opts ...ClientOpt) (*Client, error) {
	if addr == "" {
		addr = "http://localhost:8080/"
		if apiKey == "" {
			if u, err := user.Current(); err == nil {
				apiKey = u.Username
				if apiKey == "" {
					apiKey = u.Name
				}
			}
		}
	}

	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	c := &Client{
		url:          u,
		httpClient:   http.DefaultClient,
		pollInterval: time.Second / 2,
		bufferSize:   8 * 1024,
	}
	c.httpClient.Transport = http.DefaultTransport
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	c.httpClient.Transport = &RoundTripper{c.httpClient.Transport, apiKey}

	return c, nil
}

func (c *Client) get(url *url.URL, i interface{}) error {
	res, err := c.httpClient.Get(url.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if err = c.err(res); err != nil {
		return err
	}

	return json.NewDecoder(res.Body).Decode(i)
}

func (c *Client) post(url *url.URL, r io.Reader, contentType string, i interface{}) error {
	res, err := c.httpClient.Post(url.String(), contentType, r)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if err = c.err(res); err != nil {
		return err
	}

	return json.NewDecoder(res.Body).Decode(i)
}

func (c *Client) err(res *http.Response) error {
	if res.StatusCode < 299 && res.StatusCode >= 200 {
		return nil
	}

	e := &pb.Error{}
	if err := json.NewDecoder(res.Body).Decode(e); err != nil && e.Message != "" {
		return fmt.Errorf("HTTP %d: unable to parse message", res.StatusCode)
	}

	e.HTTPStatusCode = res.StatusCode
	return e
}
