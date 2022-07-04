package rototiller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/bufbuild/connect-go"
	errv1 "github.com/logsquaredn/rototiller/api/err/v1"
	"github.com/logsquaredn/rototiller/api/storage/v1/storagev1connect"
)

func NewClient(rawBaseURL, apiKey string, opts ...ClientOpt) (*Client, error) {
	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, err
	}

	c := &Client{
		baseURL:      baseURL,
		httpClient:   http.DefaultClient,
		pollInterval: time.Second / 2,
		bufferSize:   8 * 1024,
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.httpClient.Jar, err = cookiejar.New(&cookiejar.Options{}); err != nil {
		return nil, err
	}
	c.httpClient.Jar.SetCookies(baseURL, []*http.Cookie{
		{
			Name:  APIKeyCookie,
			Value: apiKey,
		},
	})

	c.storageClient = storagev1connect.NewStorageServiceClient(
		c.httpClient,
		c.baseURL.String(),
		connect.WithSendGzip(),
	)

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

	e := &errv1.Error{}
	if err := json.NewDecoder(res.Body).Decode(e); err != nil && e.Message != "" {
		return fmt.Errorf("HTTP %d: unable to parse message", res.StatusCode)
	}

	return fmt.Errorf("HTTP %d: %s", res.StatusCode, e.Message)
}
