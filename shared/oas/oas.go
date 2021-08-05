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
	svc    *s3.S3
	upldr  *s3manager.Uploader
	dwnldr *s3manager.Downloader
	bucket string
	prefix string
}

func New(sess *session.Session, bucket string, opts ...OasOpt) (*Oas, error) {
	if sess == nil {
		return nil, fmt.Errorf("oas: nil session")
	}

	if bucket == "" {
		return nil, fmt.Errorf("oas: empty bucket")
	}

	o := &Oas{}
	for _, opt := range opts {
		opt(o)
	}

	if o.prefix == "" {
		o.prefix = "jobs"
	}

	o.svc = s3.New(sess)
	o.upldr = s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.S3 = o.svc
	})
	o.dwnldr = s3manager.NewDownloader(sess, func(d *s3manager.Downloader) {
		d.S3 = o.svc
	})
	o.bucket = bucket

	return o, nil
}

const (
	input  = "input"
	output = "output"
)

func (o *Oas) PutJobInput(id string, body io.Reader, ext string) (*s3manager.UploadOutput, error) {
	prefix := path.Join(o.prefix, id, input, fmt.Sprintf("%s.%s", input, ext))
	return o.putObjects(prefix, body)
}

func (o *Oas) PutJobOutput(id string, body io.Reader, ext string) (*s3manager.UploadOutput, error) {
	prefix := path.Join(o.prefix, id, output, fmt.Sprintf("%s.%s", output, ext))
	return o.putObjects(prefix, body)
}

func (o *Oas) putObjects(key string, body io.Reader) (uploadOutput *s3manager.UploadOutput, err error) {
	uploadOutput, err = o.upldr.Upload(&s3manager.UploadInput{
		Body:   body,
		Bucket: &o.bucket,
		Key:    &key,
	})

	return
}

func (o *Oas) GetJobOutput(id string, body io.WriterAt, ext string) (err error) {
	prefix := path.Join(o.prefix, id, output, fmt.Sprintf("%s.%s", output, ext))
	_, err = o.dwnldr.Download(body, &s3.GetObjectInput{
		Bucket: &o.bucket,
		Key:    &prefix,
	})

	return
}
