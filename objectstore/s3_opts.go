package objectstore

type S3ObjectstoreOpts struct {
	Bucket          string
	Prefix          string
	Endpoint        string
	DisableSSL      bool
	ForcePathStyle  bool
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}
