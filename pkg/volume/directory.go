package volume

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func NewDir(path string) (Volume, error) {
	return Directory(path), os.MkdirAll(path, 0o755)
}

type DirectoryFile struct {
	*os.File
	Path string
}

func (f *DirectoryFile) GetName() string {
	return strings.TrimPrefix(f.Name(), f.Path)
}

func (f *DirectoryFile) GetSize() int {
	i, _ := f.Stat()
	return int(i.Size())
}

type Directory string

func (v Directory) Walk(fn WalkVolFunc) error {
	return filepath.WalkDir(string(v), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		return fn(
			string(v),
			&DirectoryFile{
				File: file,
				Path: string(v),
			},
			err,
		)
	})
}

func (v Directory) Download(path string) error {
	return fmt.Errorf("github.com/logsquaredn/rototiller/pkg/volume.*Directory.Download not implemented")
}
