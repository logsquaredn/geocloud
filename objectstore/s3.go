package objectstore

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/logsquaredn/geocloud"
	"github.com/tedsuo/ifrit"
)

type S3Objectstore struct {
	Bucket         string `long:"bucket" description:"S3 bucket"`
	Prefix         string `long:"prefix" default:"jobs" description:"Prefix to apply to keys"`
	Endpoint       string `long:"endpoint" description:"Endpoint to target"`
	DisableSSL     bool   `long:"disable-ssl" description:"Disable SSL"`
	ForcePathStyle bool   `long:"force-path-style" description:"Force S3 path style"`

	cfg    *aws.Config
	svc    *s3.S3
	upldr  *s3manager.Uploader
	dwnldr *s3manager.Downloader
}

var _ geocloud.Objectstore = (*S3Objectstore)(nil)
var _ geocloud.AWSComponent = (*S3Objectstore)(nil)

func (s *S3Objectstore) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	cfg := aws.NewConfig().WithDisableSSL(s.DisableSSL).WithS3ForcePathStyle(s.ForcePathStyle)
	if s.Endpoint != "" {
		cfg.WithEndpoint(s.Endpoint)
	}
	sess, err := session.NewSession(s.cfg, cfg)
	if err != nil {
		return err
	}
	s.svc = s3.New(sess)
	s.upldr = s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.S3 = s.svc
	})
	s.dwnldr = s3manager.NewDownloader(sess, func(d *s3manager.Downloader) {
		d.S3 = s.svc
	})

	close(ready)
	<-signals
	return nil
}

func (s *S3Objectstore) Execute(_ []string) error {
	return <-ifrit.Invoke(s).Wait()
}

func (s *S3Objectstore) Name() string {
	return "s3"
}

func (s *S3Objectstore) IsConfigured() bool {
	return s != nil && s.cfg != nil && s.Bucket != ""
}

func (s *S3Objectstore) WithConfig(cfg *aws.Config) geocloud.AWSComponent {
	s.cfg = cfg
	return s
}

func (s *S3Objectstore) GetInput(m geocloud.Message) (geocloud.Volume, error) {
	prefix := filepath.Join(s.Prefix, m.ID(), "input")
	o, err := s.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: &s.Bucket,
		Prefix: &prefix,
	})
	if err != nil {
		return nil, fmt.Errorf("error listing objects: %w", err)
	} else if len(o.Contents) > 1 {
		return nil, fmt.Errorf("multiple inputs found")
	} else if len(o.Contents) == 0 {
		return nil, fmt.Errorf("zero inputs found")
	}

	return &s3Volume{
		objs: o.Contents,
		bucket: s.Bucket,
		prefix: prefix,
		dwnldr: s.dwnldr,
	}, nil
}

func (s *S3Objectstore) GetOutput(m geocloud.Message) (geocloud.Volume, error) {
	prefix := filepath.Join(s.Prefix, m.ID(), "output")
	o, err := s.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: &s.Bucket,
		Prefix: &prefix,
	})
	if err != nil {
		return nil, fmt.Errorf("error listing objects: %w", err)
	} else if len(o.Contents) == 0 {
		return nil, fmt.Errorf("zero outputs found")
	}

	return &s3Volume{
		objs: o.Contents,
		bucket: s.Bucket,
		prefix: prefix,
		dwnldr: s.dwnldr,
	}, nil
}

func (s *S3Objectstore) PutInput(m geocloud.Message, v geocloud.Volume) error {
	var objs []s3manager.BatchUploadObject
	if err := v.Walk(func(_ string, f geocloud.File, err error) error {
		key := filepath.Join(s.Prefix, m.ID(), "input", f.Name())
		objs = append(objs, s3manager.BatchUploadObject{
			Object: &s3manager.UploadInput{
				Bucket: &s.Bucket,
				Key: &key,
				Body: f,
			},
		})
		return err
	}); err != nil {
		return err
	}

	return s.upldr.UploadWithIterator(aws.BackgroundContext(), &s3manager.UploadObjectsIterator{
		Objects: objs,
	})
}

func (s *S3Objectstore) PutOutput(m geocloud.Message, v geocloud.Volume) error {
	var objs []s3manager.BatchUploadObject
	err := v.Walk(func(_ string, f geocloud.File, err error) error {
		key := filepath.Join(s.Prefix, m.ID(), "output", f.Name())
		objs = append(objs, s3manager.BatchUploadObject{
			Object: &s3manager.UploadInput{
				Bucket: &s.Bucket,
				Key: &key,
				Body: f,
			},
		})
		return err
	})
	if err != nil {
		return err
	}

	return s.upldr.UploadWithIterator(aws.BackgroundContext(), &s3manager.UploadObjectsIterator{
		Objects: objs,
	})
}
