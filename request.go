package geocloud

import (
	"bytes"
	"io"
)

type Request interface {
	io.Reader
	Query() map[string]string
}

type request struct {
	body  io.Reader
	query map[string]string
}

func (r *request) Read(p []byte) (n int, err error) {
	return r.body.Read(p)
}

func (r *request) Query() map[string]string {
	return r.query
}

func NewStorageWithName(r io.Reader, name string) Request {
	return &request{r, map[string]string{"name": name}}
}

func NewJobFromInput(r io.Reader) Request {
	return &request{body: r}
}

func NewJobWithInput(id string) Request {
	return &request{new(bytes.Reader), map[string]string{"input": id}}
}

func NewJobWithInputOfJob(id string) Request {
	return &request{new(bytes.Reader), map[string]string{"input-of": id}}
}

func NewJobWithOutputOfJob(id string) Request {
	return &request{new(bytes.Reader), map[string]string{"output-of": id}}
}
