package objectstore

import (
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/logsquaredn/geocloud"
)

type s3File struct {
	obj    *s3.Object
	bucket string
	prefix string
	dwnldr *s3manager.Downloader
}

var _ geocloud.File = (*s3File)(nil)

func (f *s3File) Name() string {
	return strings.TrimPrefix(*f.obj.Key, f.prefix)
}

func (f *s3File) Read(p []byte) (int, error) {
	w := aws.NewWriteAtBuffer(p)
	n, err := f.dwnldr.Download(w, &s3.GetObjectInput{
		Bucket: &f.bucket,
		Key: f.obj.Key,
	})
	return int(n), err
}

type s3Volume struct {
	objs   []*s3.Object
	bucket string
	prefix string
	dwnldr *s3manager.Downloader
}

var _ geocloud.Volume = (*s3Volume)(nil)

func (v *s3Volume) Walk(fn geocloud.WalkVolFunc) (err error) {
	for _, obj := range v.objs {
		file := &s3File{
			obj: obj,
			bucket: v.bucket,
			prefix: v.prefix,
			dwnldr: v.dwnldr,
		}
		err = fn(v.prefix, file, err)
	}

	return err
}

func (v *s3Volume) Download(path string) error {
	objs := make([]s3manager.BatchDownloadObject, len(v.objs))
	for i, obj := range v.objs {
		name := strings.TrimPrefix(*obj.Key, v.prefix)
		w, err := os.Create(name)
		if err != nil {
			return err
		}
		
		objs[i] = s3manager.BatchDownloadObject{
			Object: &s3.GetObjectInput {
				Bucket: &v.bucket,
				Key: obj.Key,
			},
			Writer: w,
		}
	}

	return v.dwnldr.DownloadWithIterator(aws.BackgroundContext(), &s3manager.DownloadObjectsIterator{
		Objects: objs,
	})
}
