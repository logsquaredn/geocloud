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

func (v *bytesVolume) Name() string {
	return v.name
}

func (v *bytesVolume) Read(p []byte) (int, error) {
	return v.reader.Read(p)
}

func (v *bytesVolume) Size() int {
	// not implemented
	return 0
}

func (v *bytesVolume) Walk(fn geocloud.WalkVolFunc) error {
	return fn("", v, nil)
}

func (v *bytesVolume) Download(path string) error {
	return fmt.Errorf("bytesVolume.Download not implemented")
}
