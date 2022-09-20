package httpsink

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/logsquaredn/rototiller/pkg/sink"
)

func init() {
	opener := new(URLOpener)
	sink.DefaultURLMux().RegisterSink(SchemeHTTPS, opener)
	sink.DefaultURLMux().RegisterSink(SchemeHTTP, opener)
}

const (
	SchemeHTTPS = "https"
	SchemeHTTP  = "http"
	Scheme      = SchemeHTTPS
)

type URLOpener struct{}

func (o *URLOpener) OpenSink(ctx context.Context, u *url.URL) (sink.Sink, error) {
	return OpenSink(ctx, u)
}

type Http struct {
	*http.Client
	*url.URL
}

func OpenSink(ctx context.Context, u *url.URL) (*Http, error) {
	return &Http{http.DefaultClient, u}, nil
}

func (s *Http) Drain(ctx context.Context, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	_, err = http.Post(s.String(), http.DetectContentType(data), bytes.NewReader(data))
	return err
}
