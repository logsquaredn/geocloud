package geocloud

import (
	"bytes"
	"io"
)

type Request interface {
	io.Reader
	Query() map[string]string
	ContentType() string
}

type request struct {
	body        io.Reader
	contentType string
	query       map[string]string
}

func (r *request) Read(p []byte) (n int, err error) {
	return r.body.Read(p)
}

func (r *request) Query() map[string]string {
	return r.query
}

func (r *request) ContentType() string {
	return r.contentType
}

func NewStorageWithName(r io.Reader, contentType string, name string) Request {
	return &request{r, contentType, map[string]string{"name": name}}
}

func NewJobFromInput(r io.Reader, contentType string, query map[string]string) Request {
	return &request{r, contentType, query}
}

func NewJobWithInput(id string, query map[string]string) Request {
	q := query
	q["input"] = id
	return &request{new(bytes.Reader), "", q}
}

func NewJobWithInputOfJob(id string, query map[string]string) Request {
	q := query
	q["input-of"] = id
	return &request{new(bytes.Reader), "", q}
}

func NewJobWithOutputOfJob(id string, query map[string]string) Request {
	q := query
	q["output-of"] = id
	return &request{new(bytes.Reader), "", q}
}
