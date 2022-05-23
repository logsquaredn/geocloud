package main

import (
	"github.com/logsquaredn/geocloud/objectstore"
	"github.com/spf13/viper"
)

func init() {
	_ = viper.BindEnv("s3-archive-access-key-id", "GEOCLOUD_ACCESS_KEY_ID")
	_ = viper.BindEnv("s3-archive-secret-access-key", "GEOCLOUD_SECRET_ACCESS_KEY")
	bindConfToFlags(secretaryCmd.Flags(), []*conf{
		{"s3-archive-bucket", "", "S3 bucket"},
		{"s3-archive-prefix", "", "S3 prefix"},
		{"s3-archive-endpoint", "", "S3 endpoint"},
		{"s3-archive-disable-ssl", false, "S3 disable SSL"},
		{"s3-archive-force-path-style", false, "S3 force path style"},
		{"s3-archive-region", "us-east-1", "S3 region"},
		{"s3-archive-access-key-id", "", "S3 access key ID"},
		{"s3-archive-secret-access-key", "", "S3 secret access key"},
	}...)
}

func getS3ArchiveOpts() *objectstore.S3Opts {
	s3Opts := &objectstore.S3Opts{
		Bucket:          viper.GetString("s3-archive-bucket"),
		Prefix:          viper.GetString("s3-archive-prefix"),
		Endpoint:        viper.GetString("s3-archive-endpoint"),
		DisableSSL:      viper.GetBool("s3-archive-disable-ssl"),
		ForcePathStyle:  viper.GetBool("s3-archive-force-path-style"),
		Region:          viper.GetString("s3-archive-region"),
		AccessKeyID:     viper.GetString("s3-archive-access-key-id"),
		SecretAccessKey: viper.GetString("s3-archive-secret-access-key"),
	}
	return s3Opts
}
