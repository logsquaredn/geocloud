package aggregator

import (
	"net/http"

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

func WithContainerdNamespace(namespace string) S3AggregatorOpt {
	return func(a *S3Aggregrator) {
		a.namespace = namespace
	}
}

func WithContainerdSocket(socket string) S3AggregatorOpt {
	return func(a *S3Aggregrator) {
		a.sock = socket
	}
}

func WithHttpClient(client *http.Client) S3AggregatorOpt {
	return func(a *S3Aggregrator) {
		a.hclient = client
	}
}
