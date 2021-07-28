package sharedcmd

import (
	"time"

	"github.com/jessevdk/go-flags"
)

type S3 struct {
	Bucket string `long:"bucket" description:"S3 bucket"`
	Prefix string `long:"prefix" default:"jobs" description:"S3 key prefix"`
}

type SQS struct {
	QueueNames []string `long:"queue-name" description:"SQS queue names to listen on"`
	QueueURLs  []string `long:"queue-url" description:"SQS queue urls to listen on"`
	Visibility time.Duration `long:"visibility-timeout" default:"15s" description:"Visibilty timeout for SQS messages"`
}

type AWS struct {
	AccessKeyID     string         `long:"access-key-id" env:"GEOCLOUD_ACCESS_KEY_ID" description:"AWS access key ID"`
	SecretAccessKey string         `long:"secret-access-key" env:"GEOCLOUD_SECRET_ACCESS_KEY" description:"AWS secret access key"`
	Region          string         `long:"region" default:"us-east-1" description:"AWS region"`
	Profile         string         `long:"profile" description:"AWS profile"`
	SharedCreds     flags.Filename `long:"shared-credentials-file" default:"~/.aws/credentials" description:"Path to AWS shared credentials file"`

	S3  S3  `group:"S3" namespace:"s3"`
	SQS SQS `group:"SQS" namespace:"sqs"`
}

type Postgres struct {
	Host     string `long:"host" description:"Postgres host"`
	Port     int    `long:"port" default:"5432" description:"Postgres port"`
	User     string `long:"user" default:"geocloud" description:"Postgres username"`
	Password string `long:"password" description:"Postgres password"`
	SSLMode  string `long:"ssl-mode" default:"disable" choice:"disable" description:"Postgres SSL mode"`
	Retries  int    `long:"retries" default:"5" description:"Number of times to retry connecting to Postgres"`
}
