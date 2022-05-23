package geocloud

import (
	"encoding/json"
	"net/http"
	"net/url"
)

func NewClient(rawBaseURL, apiKey string, opts ...ClientOpt) (*Client, error) {
	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, err
	}

	c := &Client{baseURL, http.DefaultClient}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	c.httpClient.Jar.SetCookies(baseURL, []*http.Cookie{
		{
			Name:  "X-API-Key",
			Value: apiKey,
		},
	})

	return c, nil
}

func (c *Client) get(url *url.URL, i interface{}) error {
	res, err := c.httpClient.Get(url.String())
	if err != nil {
		return err
	}

	return json.NewDecoder(res.Body).Decode(i)
}
