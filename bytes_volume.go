package geocloud

import (
	"bytes"
	"fmt"
	"io"
)

func NewBytesVolume(name string, content []byte) *BytesVolume {
	return &BytesVolume{name, bytes.NewReader(content)}
}

type BytesVolume struct {
	name   string
	reader io.Reader
}

var _ File = (*BytesVolume)(nil)
var _ Volume = (*BytesVolume)(nil)

func (v *BytesVolume) Name() string {
	return v.name
}

func (v *BytesVolume) Read(p []byte) (int, error) {
	return v.reader.Read(p)
}

func (v *BytesVolume) Size() int {
	// not implemented
	return 0
}

func (v *BytesVolume) Walk(fn WalkVolFunc) error {
	return fn("", v, nil)
}

func (v *BytesVolume) Download(path string) error {
	return fmt.Errorf("BytesVolume.Download not implemented")
}
