package main

import (
	"os"

	"github.com/logsquaredn/geocloud/objectstore"
)

var (
	s3ArchiveOpts = &objectstore.S3ObjectstoreOpts{}
)

func init() {
	rootCmd.PersistentFlags().StringVar(
		&s3ArchiveOpts.Bucket,
		"s3-archive-bucket",
		os.Getenv("GEOCLOUD_S3_ARCHIVE_BUCKET"),
		"S3 archive bucket",
	)
	rootCmd.PersistentFlags().StringVar(
		&s3ArchiveOpts.Prefix,
		"s3-archive-prefix",
		os.Getenv("GEOCLOUD_S3_ARCHIVE_PREFIX"),
		"S3 archive prefix",
	)
	rootCmd.PersistentFlags().StringVar(
		&s3ArchiveOpts.Endpoint,
		"s3-archive-endpoint",
		os.Getenv("GEOCLOUD_S3_ARCHIVE_ENDPOINT"),
		"S3 archive endpoint",
	)
	rootCmd.PersistentFlags().BoolVar(
		&s3ArchiveOpts.DisableSSL,
		"s3-archive-disable-ssl",
		parseBool(os.Getenv("GEOCLOUD_S3_ARCHIVE_DISABLE_SSL")),
		"S3 archive disable ssl",
	)
	rootCmd.PersistentFlags().BoolVar(
		&s3ArchiveOpts.ForcePathStyle,
		"s3-archive-force-path-style",
		parseBool(os.Getenv("GEOCLOUD_S3_ARCHIVE_FORCE_PATH_STYLE")),
		"S3 archive force path style",
	)
	rootCmd.PersistentFlags().StringVar(
		&s3ArchiveOpts.Region,
		"s3-archive-region",
		coalesceString(
			os.Getenv("GEOCLOUD_S3_ARCHIVE_REGION"),
			"us-east-1",
		),
		"S3 archive region",
	)
	rootCmd.PersistentFlags().StringVar(
		&s3ArchiveOpts.AccessKeyID,
		"s3-access-key-id-archive",
		coalesceString(
			os.Getenv("GEOCLOUD_ACCESS_KEY_ID_ARCHIVE"),
			os.Getenv("GEOCLOUD_AWS_ACCESS_KEY_ID_ARCHIVE"),
		),
		"S3 access key ID",
	)
	rootCmd.PersistentFlags().StringVar(
		&s3ArchiveOpts.SecretAccessKey,
		"s3-secret-access-key-archive",
		coalesceString(
			os.Getenv("GEOCLOUD_SECRET_ACCESS_KEY_ARCHIVE"),
			os.Getenv("GEOCLOUD_AWS_SECRET_ACCESS_KEY_ARCHIVE"),
		),
		"S3 secret access key",
	)
}

func getS3ArchiveOpts() *objectstore.S3ObjectstoreOpts {
	return s3ArchiveOpts
}
