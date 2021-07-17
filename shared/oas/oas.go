package oas

import (
	"fmt"
	"io"
	"path"

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
	o := &Oas{}
	for _, opt := range opts {
		opt(o)
	}

	if sess == nil {
		return nil, fmt.Errorf("oas: nil session")
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

const (
	input = "input"
	output = "output"
)

func (o *Oas) PutJobInput(id string, content io.Reader) (*s3manager.UploadOutput, error) {
	key := path.Join(o.prefix, id, input)
	return o.upldr.Upload(&s3manager.UploadInput{
		Bucket: &o.bucket,
		Key:    &key,
		Body:   content,
	})
}

func (o *Oas) PutJobOutput(id string, content io.Reader) (*s3manager.UploadOutput, error) {
	key := path.Join(o.prefix, id, output)
	return o.upldr.Upload(&s3manager.UploadInput{
		Bucket: &o.bucket,
		Key:    &key,
		Body:   content,
	})
}
