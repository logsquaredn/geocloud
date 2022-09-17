package client

import (
	"net/http"

	"github.com/logsquaredn/rototiller/pkg/service"
)

type RoundTripper struct {
	http.RoundTripper
	ownerID string
}

func (r *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(service.OwnerIDHeader, r.ownerID)
	return r.RoundTripper.RoundTrip(req)
}
