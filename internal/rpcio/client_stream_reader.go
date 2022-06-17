package rpcio

import (
	"bytes"
	"io"

	"github.com/bufbuild/connect-go"
)

func NewClientStreamReader[T any](stream *connect.ClientStream[T], convert func(*T) []byte) io.Reader {
	return &ClientStreamReader[T]{
		ClientStream: stream,
		Buffer:       new(bytes.Buffer),
		Convert:      convert,
	}
}

type ClientStreamReader[T any] struct {
	ClientStream *connect.ClientStream[T]
	Buffer       *bytes.Buffer
	Convert      func(*T) []byte
}

var _ io.Reader = &ClientStreamReader[any]{}

func (r *ClientStreamReader[T]) Read(p []byte) (int, error) {
	var (
		pLen = len(p)
	)
	for r.Buffer.Len() < pLen && r.ClientStream.Receive() {
		var (
			msg = r.ClientStream.Msg()
			err = r.ClientStream.Err()
		)
		if msg == nil || err != nil {
			return 0, io.ErrClosedPipe
		}

		if _, err = r.Buffer.Write(r.Convert(msg)); err != nil {
			return 0, io.ErrClosedPipe
		}
	}

	return r.Buffer.Read(p)
}
