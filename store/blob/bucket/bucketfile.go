package bucket

import "gocloud.dev/blob"

type BucketFile struct {
	Name string
	*blob.Reader
}

func (f *BucketFile) GetName() string {
	return f.Name
}

func (f *BucketFile) GetSize() int {
	return int(f.Reader.Size())
}
