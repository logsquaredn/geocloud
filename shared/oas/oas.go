package oas

import (
	"fmt"
	"io"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Oas struct {
	svc     *s3.S3
	upldr   *s3manager.Uploader
	dwnldr  *s3manager.Downloader
	bucket  string
	prefix  string
}

func New(sess *session.Session, bucket string, opts ...OasOpt) (*Oas, error) {
	if sess == nil {
		return nil, fmt.Errorf("oas: nil session")
	}

	o := &Oas{}
	for _, opt := range opts {
		opt(o)
	}

	if o.prefix == "" {
		o.prefix = "jobs"
	}

	o.svc = s3.New(sess)
	o.upldr = s3manager.NewUploader(sess)
	o.dwnldr = s3manager.NewDownloader(sess)
	o.bucket = bucket

	return o, nil
}

var ctx = aws.BackgroundContext()

const (
	input = "input"
	output = "output"
)

func (o *Oas) PutJobInput(id string, bodies ...io.Reader) error {
	prefix := path.Join(o.prefix, id, input)
	return o.putObjects(prefix, bodies...)
}

func (o *Oas) PutJobOutput(id string, bodies ...io.Reader) error {
	prefix := path.Join(o.prefix, id, output)
	return o.putObjects(prefix, bodies...)
}

func (o *Oas) putObjects(prefix string, bodies ...io.Reader) error {
	objects := make([]s3manager.BatchUploadObject, len(bodies))
	for i, body := range bodies {
		key := path.Join(prefix, fmt.Sprintf("%d", i))
		objects[i] = s3manager.BatchUploadObject{
			Object: &s3manager.UploadInput{
				Body: body,
				Bucket: &o.bucket,
				Key: &key,
			},
		}
	}

	return o.upldr.UploadWithIterator(ctx, &s3manager.UploadObjectsIterator{
		Objects: objects,
	})
}
