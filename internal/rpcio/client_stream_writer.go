package rpcio

import (
	"errors"
	"io"

	"github.com/bufbuild/connect-go"
)

func NewClientStreamWriter[T1, T2 any](stream *connect.ClientStreamForClient[T1, T2], convert func([]byte) *T1) io.Writer {
	return &ClientStreamWriter[T1, T2]{
		ClientStream: stream,
		Convert:      convert,
	}
}

type ClientStreamWriter[T1, T2 any] struct {
	ClientStream *connect.ClientStreamForClient[T1, T2]
	Convert      func([]byte) *T1
}

var _ io.Writer = &ClientStreamWriter[any, any]{}

func (w *ClientStreamWriter[T1, T2]) Write(p []byte) (int, error) {
	err := w.ClientStream.Send(w.Convert(p))
	switch {
	case errors.Is(err, io.EOF):
		_, err = w.ClientStream.CloseAndReceive()
		return 0, err
	// this should never happen according to connect (see godoc for ClientStream.Send)
	case err != nil:
		return 0, err
	}

	return len(p), nil
}
