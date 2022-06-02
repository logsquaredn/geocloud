//nolint:dupl
package main

import (
	"github.com/logsquaredn/geocloud/internal/conf"
	"github.com/logsquaredn/geocloud/objectstore"
	"github.com/spf13/viper"
)

func init() {
	_ = viper.BindEnv("s3-archive-access-key-id", "GEOCLOUD_ACCESS_KEY_ID")
	_ = viper.BindEnv("s3-archive-secret-access-key", "GEOCLOUD_SECRET_ACCESS_KEY")
	_ = conf.BindToFlags(secretaryCmd.Flags(), nil, []*conf.Conf{
		{
			Arg:         "s3-archive-bucket",
			Default:     "",
			Description: "S3 bucket",
		},
		{
			Arg:         "s3-archive-prefix",
			Default:     "",
			Description: "S3 prefix",
		},
		{
			Arg:         "s3-archive-endpoint",
			Default:     "",
			Description: "S3 endpoint",
		},
		{
			Arg:         "s3-archive-disable-ssl",
			Default:     false,
			Description: "S3 disable SSL",
		},
		{
			Arg:         "s3-archive-force-path-style",
			Default:     false,
			Description: "S3 force path style",
		},
		{
			Arg:         "s3-archive-region",
			Default:     "us-east-1",
			Description: "S3 region",
		},
		{
			Arg:         "s3-archive-access-key-id",
			Default:     "",
			Description: "S3 access key ID",
		},
		{
			Arg:         "s3-archive-secret-access-key",
			Default:     "",
			Description: "S3 secret access key",
		},
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
