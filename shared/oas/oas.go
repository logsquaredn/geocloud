package oas

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Oas struct {
	s3      *s3.S3
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
	o.s3, err = o.GetS3Service()
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (o *Oas) PutObject(key string) error {
	putObjectParams := o.s3.PutObjectInput {
		Bucket: o.bucket,
		
	}
}

func (o *Oas) GetS3Service() (*s3.S3, error) {
	if o.sess == nil {
		var err error
		o.sess, err = o.GetSession()
		if err != nil {
			return nil, err
		}
	}

	return s3.New(o.sess), nil
}

func (o *Oas) GetSession() (*session.Session, error) {
	if o.hClient == nil {
		o.hClient = http.DefaultClient
	}

	cfg := aws.NewConfig().WithHTTPClient(o.hClient).WithRegion(o.region).WithCredentials(o.creds)

	return session.NewSession(cfg)
}
