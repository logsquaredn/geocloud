package geocloud

import "io"

// File ...
type File interface {
	io.ReadCloser

	// Name returns the path to the file relative to the
	// File's Volume
	Name() string
	Size() int
}

// WalkVolFunc ...
type WalkVolFunc func(string, File, error) error

// Volume ...
type Volume interface {
	// Walk iterates over each File in the Volume,
	// calling WalkVolFunc for each one
	Walk(WalkVolFunc) error
	// Download copies the contents of each File in the Volume
	// to the directory at the given path
	Download(string) error
}
