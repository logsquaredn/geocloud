package client

import (
	"net/http"

	"github.com/logsquaredn/rototiller/api"
)

type RoundTripper struct {
	http.RoundTripper
	namespace string
}

func (r *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(api.NamespaceHeader, r.namespace)
	return r.RoundTripper.RoundTrip(req)
}
