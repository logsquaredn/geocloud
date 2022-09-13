package client

import (
	"net/http"

	"github.com/logsquaredn/rototiller/pkg/service"
)

type RoundTripper struct {
	http.RoundTripper
	apiKey string
}

func (r *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(service.APIKeyHeader, r.apiKey)
	return r.RoundTripper.RoundTrip(req)
}
