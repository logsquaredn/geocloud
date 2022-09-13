package client

import (
	"github.com/logsquaredn/rototiller/pkg/service"
	"net/http"
)

type RoundTripper struct {
	http.RoundTripper
	apiKey string
}

func (r *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(service.APIKeyHeader, r.apiKey)
	return r.RoundTripper.RoundTrip(req)
}
