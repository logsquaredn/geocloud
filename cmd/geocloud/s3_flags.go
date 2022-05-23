package main

import (
	"github.com/logsquaredn/geocloud/objectstore"
	"github.com/spf13/viper"
)

func init() {
	_ = viper.BindEnv("s3-access-key-id", "GEOCLOUD_ACCESS_KEY_ID")
	_ = viper.BindEnv("s3-secret-access-key", "GEOCLOUD_SECRET_ACCESS_KEY")
	bindConfToFlags(rootCmd.PersistentFlags(), []*conf{
		{"s3-bucket", "", "S3 bucket"},
		{"s3-prefix", "", "S3 prefix"},
		{"s3-endpoint", "", "S3 endpoint"},
		{"s3-disable-ssl", false, "S3 disable SSL"},
		{"s3-force-path-style", false, "S3 force path style"},
		{"s3-region", "us-east-1", "S3 region"},
		{"s3-access-key-id", "", "S3 access key ID"},
		{"s3-secret-access-key", "", "S3 secret access key"},
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
