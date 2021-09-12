package aggregator

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
)

func (a *S3Aggregrator) pull(ctx context.Context, ref string) (containerd.Image, error) {
	return a.cclient.Pull(
		ctx, fmt.Sprintf("%s/%s", a.host, ref),
		containerd.WithPullUnpack,
		containerd.WithResolver(*a.resolver),
	)
}
