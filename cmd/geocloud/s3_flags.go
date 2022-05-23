package main

import (
	"github.com/logsquaredn/geocloud/objectstore"
	"github.com/spf13/viper"
)

func init() {
	viper.BindEnv("s3-access-key-id", "GEOCLOUD_ACCESS_KEY_ID")
	viper.BindEnv("s3-secret-access-key", "GEOCLOUD_SECRET_ACCESS_KEY")
	bindConfToFlags(rootCmd.PersistentFlags(), []*conf{
		{
			arg:  "s3-bucket",
			def:  "",
			desc: "S3 bucket",
		},
		{
			arg:  "s3-prefix",
			def:  "",
			desc: "S3 prefix",
		},
		{
			arg:  "s3-endpoint",
			def:  "",
			desc: "S3 endpoint",
		},
		{
			arg:  "s3-disable-ssl",
			def:  false,
			desc: "S3 disable SSL",
		},
		{
			arg:  "s3-force-path-style",
			def:  false,
			desc: "S3 force path style",
		},
		{
			arg:  "s3-region",
			def:  "us-east-1",
			desc: "S3 region",
		},
		{
			arg:  "s3-access-key-id",
			def:  "",
			desc: "S3 access key ID",
		},
		{
			arg:  "s3-secret-access-key",
			def:  "",
			desc: "S3 secret access key",
		},
	}...)
}

func getS3Opts() *objectstore.S3Opts {
	s3Opts := &objectstore.S3Opts{
		Bucket:          viper.GetString("s3-bucket"),
		Prefix:          viper.GetString("s3-prefix"),
		Endpoint:        viper.GetString("s3-endpoint"),
		DisableSSL:      viper.GetBool("s3-disable-ssl"),
		ForcePathStyle:  viper.GetBool("s3-force-path-style"),
		Region:          viper.GetString("s3-region"),
		AccessKeyID:     viper.GetString("s3-access-key-id"),
		SecretAccessKey: viper.GetString("s3-secret-access-key"),
	}
	return s3Opts
}
