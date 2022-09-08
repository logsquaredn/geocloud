package bucket

import (
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/logsquaredn/rototiller/pkg/volume"
	"gocloud.dev/blob"
	"golang.org/x/sync/errgroup"
)

func New(ctx context.Context, addr string) (*Blobstore, error) {
	if addr == "" {
		addr = os.Getenv("S3_BUCKET")
	}

	addr = "s3://" + strings.TrimPrefix(addr, "s3://")

	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	q := u.Query()

	for queryParam, envVar := range map[string]string{
		"disableSSL":       "S3_DISABLE_SSL",
		"s3ForcePathStyle": "S3_FORCE_PATH_STYLE",
		"endpoint":         "S3_ENDPOINT",
	} {
		if value := os.Getenv(envVar); value != "" {
			q.Add(queryParam, value)
		}
	}

	u.RawQuery = q.Encode()

	bucket, err := blob.OpenBucket(ctx, u.String())
	if err != nil {
		return nil, err
	}

	return &Blobstore{bucket}, nil
}

type Blobstore struct {
	*blob.Bucket
}

func (b *Blobstore) GetObject(ctx context.Context, id string) (volume.Volume, error) {
	var (
		li    = b.List(&blob.ListOptions{Prefix: id})
		files []volume.File
		err   error
	)

	for lo, err := li.Next(ctx); err == nil; lo, err = li.Next(ctx) {
		if !lo.IsDir {
			r, err := b.NewReader(ctx, lo.Key, &blob.ReaderOptions{})
			if err != nil {
				return nil, err
			}

			files = append(files, &BucketFile{strings.TrimPrefix(lo.Key, id), r})
		}
	}
	switch {
	case errors.Is(err, io.EOF):
	case err != nil:
		return nil, err
	}

	return volume.New(files...), nil
}

func (b *Blobstore) PutObject(ctx context.Context, id string, vol volume.Volume) error {
	eg, ctx := errgroup.WithContext(ctx)

	if err := vol.Walk(func(s string, f volume.File, err error) error {
		eg.Go(func() error {
			w, err := b.NewWriter(ctx, filepath.Join(id, s, f.GetName()), &blob.WriterOptions{})
			if err != nil {
				return err
			}
			defer w.Close()

			if _, err = io.Copy(w, f); err != nil {
				return err
			}

			return w.Close()
		})

		return nil
	}); err != nil {
		return err
	}

	return eg.Wait()
}

func (b *Blobstore) DeleteObject(ctx context.Context, id string) error {
	var (
		li  = b.List(&blob.ListOptions{Prefix: id})
		err error
	)

	for lo, err := li.Next(ctx); err == nil; lo, err = li.Next(ctx) {
		if !lo.IsDir {
			if err = b.Delete(ctx, lo.Key); err != nil {
				return err
			}
		}
	}

	return err
}
