package oas

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

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

func (o *Oas) DownloadJobInputToDir(id, dir string) (*os.File, error) {
	prefix := filepath.Join(o.prefix, id, "input")
	output, err := o.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: &o.bucket,
		Prefix: &prefix,
	})
	if err != nil {
		return nil, fmt.Errorf("oas: error listing objects: %w", err)
	} else if len(output.Contents) > 1 {
		return nil, fmt.Errorf("oas: multiple job inputs found")
	} else if len(output.Contents) == 0 {
		return nil, fmt.Errorf("oas: zero job inputs found")
	}

	content := output.Contents[0]
	w, err := os.Create(filepath.Join(dir, fmt.Sprintf("input%s", filepath.Ext(*content.Key))))
	if err != nil {
		return nil, fmt.Errorf("oas: error creating file: %w", err)
	}

	_, err = o.dwnldr.Download(w, &s3.GetObjectInput{
		Bucket: &o.bucket,
		Key: content.Key,
	})
	if err != nil {
		return nil, fmt.Errorf("oas: error downloading object: %w", err)
	}

	return w, nil
}

func (o *Oas) PutJobInput(id string, body io.Reader, ext string) (*s3manager.UploadOutput, error) {
	prefix := filepath.Join(o.prefix, id, "input", fmt.Sprintf("%s.%s", "input", ext))
	return o.putObjects(prefix, body)
}

func (o *Oas) PutJobOutput(id string, body io.Reader, ext string) (*s3manager.UploadOutput, error) {
	prefix := filepath.Join(o.prefix, id, "output", fmt.Sprintf("%s.%s", "output", ext))
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
	prefix := filepath.Join(o.prefix, id, "output", fmt.Sprintf("output.%s", ext))
	_, err = o.dwnldr.Download(body, &s3.GetObjectInput{
		Bucket: &o.bucket,
		Key:    &prefix,
	})

	return
}
