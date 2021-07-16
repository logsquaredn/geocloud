package oas

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

type OasOpt func(o *Oas)

func WithSession(session *session.Session) OasOpt {
	return func(o *Oas) {
		o.sess = session
	}
}

func WithRegion(region string) OasOpt {
	return func(o *Oas) {
		o.region = region
	}
}

func WithCredentials(creds *credentials.Credentials) OasOpt {
	return func(o *Oas) {
		o.creds = creds
	}
}

func WithHttpClient(client *http.Client) OasOpt {
	return func(o *Oas) {
		o.hClient = client
	}
}

func WithBucket(bucket string) OasOpt {
	return func(o *Oas) {
		o.bucket = bucket
	}
}
