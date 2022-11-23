package api

import (
	"fmt"
	"net/http"
)

func NewErr(err error, codes ...int) *Error {
	if err == nil {
		return nil
	} else if e, ok := err.(*Error); ok {
		return e
	}

	e := &Error{
		Message:        err.Error(),
		HTTPStatusCode: http.StatusInternalServerError,
	}
	if len(codes) > 0 {
		e.HTTPStatusCode = codes[0]
	}

	return e
}

type Error struct {
	Message        string `json:"error,omitempty"`
	HTTPStatusCode int    `json:"-"`
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	return "HTTP " + fmt.Sprint(e.HTTPStatusCode) + ": " + e.Message
}
