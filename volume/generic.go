package volume

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"
)

func NewFile(name string, r io.Reader, size int) *GenericFile {
	if rc, ok := r.(io.ReadCloser); ok {
		return &GenericFile{name, rc, size}
	}

	return &GenericFile{name, io.NopCloser(r), size}
}

type GenericFile struct {
	Name string
	io.ReadCloser
	Size int
}

func (f *GenericFile) GetName() string {
	return f.Name
}

func (f *GenericFile) GetSize() int {
	return f.Size
}

func New(files ...File) *Generic {
	return &Generic{files}
}

type Generic struct {
	Files []File
}

func (v *Generic) Walk(wvf WalkVolFunc) error {
	for _, file := range v.Files {
		if err := wvf("", file, nil); err != nil {
			return err
		}
	}

	return nil
}

func (v *Generic) Download(path string) error {
	eg, _ := errgroup.WithContext(context.TODO())

	for _, f := range v.Files {
		var (
			name = f.GetName()
			src  = f
		)
		defer src.Close()

		eg.Go(func() error {
			file, err := os.Create(filepath.Join(path, name))
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(file, src)
			return err
		})
	}

	return eg.Wait()
}
