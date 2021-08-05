package aggregator

import (
	"strings"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/remotes"
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

func WithPrefetch(prefetch bool) S3AggregatorOpt {
	return func(a *S3Aggregrator) {
		a.prefetch = prefetch
	}
}

func WithRegistryHost(host string) S3AggregatorOpt {
	return func(a *S3Aggregrator) {
		a.host = strings.Trim(host, "/")
	}
}

func WithResolver(resolver *remotes.Resolver) S3AggregatorOpt {
	return func(a *S3Aggregrator) {
		a.resolver = resolver
	}
}

func WithTasks(tasks... string) S3AggregatorOpt {
	return func(a *S3Aggregrator) {
		a.tasks = append(a.tasks, tasks...)
	}
}

func WithWorkdir(dir string) S3AggregatorOpt {
	return func(a *S3Aggregrator) {
		a.workdir = dir
	}
}
