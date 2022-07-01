package errv1

import (
	"net/http"

	"github.com/bufbuild/connect-go"
)

func New(err error, codes ...int) *Error {
	if err == nil {
		return nil
	} else if e, ok := err.(*Error); ok {
		return e
	}

	e := &Error{
		Message:        err.Error(),
		HTTPStatusCode: http.StatusInternalServerError,
		ConnectCode:    connect.CodeUnknown,
	}
	switch len(codes) {
	case 2:
		e.HTTPStatusCode = codes[0]
		e.ConnectCode = connect.Code(codes[1])
	case 1:
		e.HTTPStatusCode = codes[0]
	}

	return e
}

type Error struct {
	Message        string       `json:"error,omitempty"`
	HTTPStatusCode int          `json:"-"`
	ConnectCode    connect.Code `json:"-"`
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	return e.Message
}
