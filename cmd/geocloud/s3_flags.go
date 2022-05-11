package main

import (
	"os"

	"github.com/logsquaredn/geocloud/objectstore"
)

var (
	s3Opts = &objectstore.S3ObjectstoreOpts{}
)

func init() {
	rootCmd.PersistentFlags().StringVar(
		&s3Opts.Bucket,
		"s3-bucket",
		os.Getenv("GEOCLOUD_S3_BUCKET"),
		"S3 bucket",
	)
	rootCmd.PersistentFlags().StringVar(
		&s3Opts.Prefix,
		"s3-prefix",
		os.Getenv("GEOCLOUD_S3_PREFIX"),
		"S3 prefix",
	)
	rootCmd.PersistentFlags().StringVar(
		&s3Opts.Endpoint,
		"s3-endpoint",
		os.Getenv("GEOCLOUD_S3_ENDPOINT"),
		"S3 endpoint",
	)
	rootCmd.PersistentFlags().BoolVar(
		&s3Opts.DisableSSL,
		"s3-disable-ssl",
		parseBool(os.Getenv("GEOCLOUD_S3_DISABLE_SSL")),
		"S3 disable ssl",
	)
	rootCmd.PersistentFlags().BoolVar(
		&s3Opts.ForcePathStyle,
		"s3-force-path-style",
		parseBool(os.Getenv("GEOCLOUD_S3_FORCE_PATH_STYLE")),
		"S3 force path style",
	)
	rootCmd.PersistentFlags().StringVar(
		&s3Opts.Region,
		"s3-region",
		coalesceString(
			os.Getenv("GEOCLOUD_S3_REGION"),
			"us-east-1",
		),
		"S3 region",
	)
	rootCmd.PersistentFlags().StringVar(
		&s3Opts.AccessKeyID,
		"s3-access-key-id",
		coalesceString(
			os.Getenv("GEOCLOUD_ACCESS_KEY_ID"),
			os.Getenv("GEOCLOUD_AWS_ACCESS_KEY_ID"),
		),
		"S3 access key ID",
	)
	rootCmd.PersistentFlags().StringVar(
		&s3Opts.SecretAccessKey,
		"s3-secret-access-key",
		coalesceString(
			os.Getenv("GEOCLOUD_SECRET_ACCESS_KEY"),
			os.Getenv("GEOCLOUD_AWS_SECRET_ACCESS_KEY"),
		),
		"S3 secret access key",
	)
}

func getS3Opts() *objectstore.S3ObjectstoreOpts {
	return s3Opts
}
