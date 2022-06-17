package rpcio

import (
	"io"

	"github.com/bufbuild/connect-go"
)

func NewServerStreamWriter[T any](stream *connect.ServerStream[T], convert func([]byte) *T) io.Writer {
	return &ServerStreamWriter[T]{
		ServerStream: stream,
		Convert:      convert,
	}
}

type ServerStreamWriter[T any] struct {
	ServerStream *connect.ServerStream[T]
	Convert      func([]byte) *T
}

var _ io.Writer = &ServerStreamWriter[any]{}

func (w *ServerStreamWriter[T]) Write(p []byte) (int, error) {
	return len(p), w.ServerStream.Send(w.Convert(p))
}
