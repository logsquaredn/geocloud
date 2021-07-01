package aggregator

import (
	"database/sql"
	"net/http"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/containerd/containerd"
)

type S3AggregatorOpt func(a *S3Aggregrator)

func WithAddress(address string) S3AggregatorOpt {
	return func(a *S3Aggregrator) {
		a.addr = address
	}
}

func WithContainerdClient(client *containerd.Client) S3AggregatorOpt {
	return func(a *S3Aggregrator) {
		a.cclient = client
	}
}

func WithCredentials(creds *credentials.Credentials) S3AggregatorOpt {
	return func(a *S3Aggregrator) {
		a.creds = creds
	}
}

func WithDB(db *sql.DB) S3AggregatorOpt {
	return func(a *S3Aggregrator) {
		a.db = db
	}
}

func WithHttpClient(client *http.Client) S3AggregatorOpt {
	return func(a *S3Aggregrator) {
		a.hclient = client
	}
}

func WithRegion(region string) S3AggregatorOpt {
	return func(a *S3Aggregrator) {
		a.region = region
	}
}

func WithService(service *s3.S3) S3AggregatorOpt {
	return func(a *S3Aggregrator) {
		a.svc = service
	}
}

func WithSession(session *session.Session) S3AggregatorOpt {
	return func(a *S3Aggregrator) {
		a.sess = session
	}
}

func WithContainerdSocket(socket string) S3AggregatorOpt {
	return func(a *S3Aggregrator) {
		a.sock = socket
	}
}
