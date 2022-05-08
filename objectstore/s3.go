package objectstore

import (
	"fmt"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/logsquaredn/geocloud"
)

type s3Objectstore struct {
	prefix string
	bucket string

	svc    *s3.S3
	upldr  *s3manager.Uploader
	dwnldr *s3manager.Downloader
}

var _ geocloud.Objectstore = (*s3Objectstore)(nil)

func NewS3(opts *S3ObjectstoreOpts) (*s3Objectstore, error) {
	var (
		s = &s3Objectstore{
			prefix: opts.Prefix,
			bucket: opts.Bucket,
		}
		creds = credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.StaticProvider{
					Value: credentials.Value{
						AccessKeyID:     opts.AccessKeyID,
						SecretAccessKey: opts.SecretAccessKey,
					},
				},
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{
					Filename: "~/.aws/creds",
					Profile:  "default",
				},
			},
		)
		cfg = aws.NewConfig().
			WithDisableSSL(opts.DisableSSL).
			WithS3ForcePathStyle(opts.ForcePathStyle).
			WithRegion(opts.Region).
			WithCredentials(creds)
	)
	if s.bucket == "" {
		return nil, fmt.Errorf("Bucket is required")
	}

	if opts.Endpoint != "" {
		cfg.WithEndpoint(opts.Endpoint)
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}

	s.svc = s3.New(sess)
	s.upldr = s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.S3 = s.svc
	})
	s.dwnldr = s3manager.NewDownloader(sess, func(d *s3manager.Downloader) {
		d.S3 = s.svc
	})

	return s, nil
}

func (s *s3Objectstore) GetInput(m geocloud.Message) (geocloud.Volume, error) {
	prefix := filepath.Join(s.prefix, m.GetID(), "input")
	o, err := s.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: &s.bucket,
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
		objs:   o.Contents,
		bucket: s.bucket,
		prefix: prefix,
		dwnldr: s.dwnldr,
	}, nil
}

func (s *s3Objectstore) GetOutput(m geocloud.Message) (geocloud.Volume, error) {
	prefix := filepath.Join(s.prefix, m.GetID(), "output")
	o, err := s.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: &s.bucket,
		Prefix: &prefix,
	})
	if err != nil {
		return nil, fmt.Errorf("error listing objects: %w", err)
	} else if len(o.Contents) == 0 {
		return nil, fmt.Errorf("zero outputs found")
	}

	return &s3Volume{
		objs:   o.Contents,
		bucket: s.bucket,
		prefix: prefix,
		dwnldr: s.dwnldr,
	}, nil
}

func (s *s3Objectstore) PutInput(m geocloud.Message, v geocloud.Volume) error {
	var objs []s3manager.BatchUploadObject
	if err := v.Walk(func(_ string, f geocloud.File, err error) error {
		key := filepath.Join(s.prefix, m.GetID(), "input", f.Name())
		objs = append(objs, s3manager.BatchUploadObject{
			Object: &s3manager.UploadInput{
				Bucket: &s.bucket,
				Key:    &key,
				Body:   f,
			},
		})
		return err
	}); err != nil {
		return err
	}

	if len(objs) == 0 {
		return fmt.Errorf("no inputs found")
	}

	return s.upldr.UploadWithIterator(aws.BackgroundContext(), &s3manager.UploadObjectsIterator{
		Objects: objs,
	})
}

func (s *s3Objectstore) PutOutput(m geocloud.Message, v geocloud.Volume) error {
	var objs []s3manager.BatchUploadObject
	err := v.Walk(func(_ string, f geocloud.File, err error) error {
		key := filepath.Join(s.prefix, m.GetID(), "output", f.Name())
		objs = append(objs, s3manager.BatchUploadObject{
			Object: &s3manager.UploadInput{
				Bucket: &s.bucket,
				Key:    &key,
				Body:   f,
			},
		})
		return err
	})
	if err != nil {
		return err
	}

	if len(objs) == 0 {
		return fmt.Errorf("no outputs found")
	}

	return s.upldr.UploadWithIterator(aws.BackgroundContext(), &s3manager.UploadObjectsIterator{
		Objects: objs,
	})
}

func (s *s3Objectstore) DeleteRecursive(prefix string) error {
	o, err := s.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: &s.bucket,
		Prefix: &prefix,
	})
	if err != nil {
		return err
	}

	for _, s3Obj := range o.Contents {
		_, err = s.svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: &s.bucket,
			Key:    s3Obj.Key,
		})
		if err != nil {
			return err
		}
	}

	_, err = s.svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &prefix,
	})
	return err
}
