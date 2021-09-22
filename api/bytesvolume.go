package api

import (
	"fmt"
	"io"

	"github.com/logsquaredn/geocloud"
)


type bytesVolume struct {
	reader io.Reader
	name   string
}

var _ geocloud.File = (*bytesVolume)(nil)
var _ geocloud.Volume = (*bytesVolume)(nil)

func (f *bytesVolume) Name() string {
	return f.name
}

func (f *bytesVolume) Read(p []byte) (int, error) {
	return f.reader.Read(p)
}

func (v *bytesVolume) Walk(fn geocloud.WalkVolFunc) error {
	return fn("", v, nil)
}

func (v *bytesVolume) Download(path string) error {
	return fmt.Errorf("bytesVolume.Download not implemented")
}
