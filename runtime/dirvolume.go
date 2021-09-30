package runtime

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/logsquaredn/geocloud"
)

type dirFile struct {
	file    *os.File
	volPath string
}

var _ geocloud.File = (*dirFile)(nil)

func (f *dirFile) Name() string {
	return strings.TrimPrefix(f.file.Name(), f.volPath)
}

func (f *dirFile) Read(p []byte) (int, error) {
	return f.file.Read(p)
}

func (f *dirFile) Size() int {
	i, _ := f.file.Stat()
	return int(i.Size())
}

type dirVolume struct {
	path string
}

var _ geocloud.Volume = (*dirVolume)(nil)

func (v *dirVolume) Walk(fn geocloud.WalkVolFunc) error {
	return filepath.WalkDir(v.path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			return nil
		}

		file, err := os.Open(filepath.Join(v.path, d.Name()))
		if err != nil {
			return err
		}

		return fn(
			v.path,
			&dirFile{
				file: file,
				volPath: v.path,
			},
			err,
		)
	})
}

func (v *dirVolume) Download(path string) error {
	return fmt.Errorf("dirVolume.Download not implemented")
}
