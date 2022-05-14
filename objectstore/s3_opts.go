package objectstore

type S3Opts struct {
	Bucket          string
	Prefix          string
	Endpoint        string
	DisableSSL      bool
	ForcePathStyle  bool
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}
