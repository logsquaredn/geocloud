package sink

import (
	"context"
	"errors"
	"io"
	"net/url"
)

var (
	ErrNoSinkRegistered = errors.New("no sink registered for the given scheme")
)

type Sink interface {
	Drain(context.Context, io.Reader) error
}

type URLOpener interface {
	OpenSink(context.Context, *url.URL) (Sink, error)
}

type URLMux struct {
	URLOpeners map[string]URLOpener
}

func (m *URLMux) RegisterSink(scheme string, opener URLOpener) {
	m.URLOpeners[scheme] = opener
}

var (
	defaultURLMux = &URLMux{
		URLOpeners: map[string]URLOpener{},
	}
)

func DefaultURLMux() *URLMux {
	return defaultURLMux
}

func OpenSink(ctx context.Context, addr string) (Sink, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	if opener, ok := defaultURLMux.URLOpeners[u.Scheme]; ok {
		return opener.OpenSink(ctx, u)
	}

	return nil, ErrNoSinkRegistered
}
