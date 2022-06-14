package geocloud

import (
	"fmt"
	"io"
)

func NewSingleFileVolume(name string, r io.Reader) Volume {
	return &singleFileVolume{name, r}
}

type singleFileVolume struct {
	name   string
	reader io.Reader
}

var _ File = (*singleFileVolume)(nil)
var _ Volume = (*singleFileVolume)(nil)

func (v *singleFileVolume) Name() string {
	return v.name
}

func (v *singleFileVolume) Read(p []byte) (int, error) {
	return v.reader.Read(p)
}

func (v *singleFileVolume) Close() error {
	// not implemented
	return nil
}

func (v *singleFileVolume) Size() int {
	// not implemented
	return 0
}

func (v *singleFileVolume) Walk(fn WalkVolFunc) error {
	return fn("", v, nil)
}

func (v *singleFileVolume) Download(path string) error {
	return fmt.Errorf("singleFileVolume.Download not implemented")
}
