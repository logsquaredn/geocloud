//nolint:dupl
package main

import (
	"github.com/logsquaredn/geocloud/internal/conf"
	"github.com/logsquaredn/geocloud/objectstore"
	"github.com/spf13/viper"
)

func init() {
	_ = viper.BindEnv("s3-access-key-id", envVarAccessKeyID)
	_ = viper.BindEnv("s3-secret-access-key", envVarSecretAccessKey)
	_ = conf.BindToFlags(rootCmd.PersistentFlags(), nil, []*conf.Conf{
		{
			Arg:         "s3-bucket",
			Default:     "",
			Description: "S3 bucket",
		},
		{
			Arg:         "s3-prefix",
			Default:     "",
			Description: "S3 prefix",
		},
		{
			Arg:         "s3-endpoint",
			Default:     "",
			Description: "S3 endpoint",
		},
		{
			Arg:         "s3-disable-ssl",
			Default:     false,
			Description: "S3 disable SSL",
		},
		{
			Arg:         "s3-force-path-style",
			Default:     false,
			Description: "S3 force path style",
		},
		{
			Arg:         "s3-region",
			Default:     "us-east-1",
			Description: "S3 region",
		},
		{
			Arg:         "s3-access-key-id",
			Default:     "",
			Description: "S3 access key ID",
		},
		{
			Arg:         "s3-secret-access-key",
			Default:     "",
			Description: "S3 secret access key",
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
