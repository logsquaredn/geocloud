package oas

import (
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Oas struct {
	upldr   *s3manager.Uploader
	sess    *session.Session
	region  string
	creds   *credentials.Credentials
	hClient *http.Client
	bucket  string
}

func New(opts ...OasOpt) (*Oas, error) {
	o := &Oas{}
	for _, opt := range opts {
		opt(o)
	}

	var err error
	o.upldr, err = o.GetUploader()
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (o *Oas) PutObject(key string, content io.Reader) (*s3manager.UploadOutput, error) {
	upParams := &s3manager.UploadInput{
		Bucket: &o.bucket,
		Key:    &key,
		Body:   content,
	}

	return o.upldr.Upload(upParams)
}

func (o *Oas) GetUploader() (*s3manager.Uploader, error) {
	if o.sess == nil {
		var err error
		o.sess, err = o.GetSession()
		if err != nil {
			return nil, err
		}
	}

	return s3manager.NewUploader(o.sess), nil
}

func (o *Oas) GetSession() (*session.Session, error) {
	if o.hClient == nil {
		o.hClient = http.DefaultClient
	}

	cfg := aws.NewConfig().WithHTTPClient(o.hClient).WithRegion(o.region).WithCredentials(o.creds)

	return session.NewSession(cfg)
}
