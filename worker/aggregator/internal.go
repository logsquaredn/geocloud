package aggregator

import (
	"fmt"

	"github.com/containerd/containerd"
)

func (a *S3Aggregrator) pull(ref string) (containerd.Image, error) {
	return a.cclient.Pull(
		ctx, fmt.Sprintf("%s/%s", a.host, ref),
		containerd.WithPullUnpack,
		containerd.WithResolver(*a.resolver),
	)
}
