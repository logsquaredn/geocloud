package client

import (
	"net/http"

	"github.com/logsquaredn/rototiller/api"
)

type RoundTripper struct {
	http.RoundTripper
	ownerID string
}

func (r *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(api.OwnerIDHeader, r.ownerID)
	return r.RoundTripper.RoundTrip(req)
}
