package main

import (
	"github.com/logsquaredn/geocloud/objectstore"
	"github.com/spf13/viper"
)

func init() {
	viper.BindEnv("s3-archive-access-key-id", "GEOCLOUD_ACCESS_KEY_ID")
	viper.BindEnv("s3-archive-secret-access-key", "GEOCLOUD_SECRET_ACCESS_KEY")
	bindConfToFlags(secretaryCmd.Flags(), []*conf{
		{
			arg:  "s3-archive-bucket",
			def:  "",
			desc: "S3 bucket",
		},
		{
			arg:  "s3-archive-prefix",
			def:  "",
			desc: "S3 prefix",
		},
		{
			arg:  "s3-archive-endpoint",
			def:  "",
			desc: "S3 endpoint",
		},
		{
			arg:  "s3-archive-disable-ssl",
			def:  "",
			desc: "S3 disable SSL",
		},
		{
			arg:  "s3-archive-force-path-style",
			def:  "",
			desc: "S3 force path style",
		},
		{
			arg:  "s3-archive-region",
			def:  "us-east-1",
			desc: "S3 region",
		},
		{
			arg:  "s3-archive-access-key-id",
			def:  "",
			desc: "S3 access key ID",
		},
		{
			arg:  "s3-archive-secret-access-key",
			def:  "",
			desc: "S3 secret access key",
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
