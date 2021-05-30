package workercmd

import (
	"context"
	"os"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/tedsuo/ifrit"
)

func (cmd *WorkerCmd) worker() (ifrit.Runner, error) {
	return ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
		var (
			address = string(cmd.Containerd.Address)
			ctx     = namespaces.WithNamespace(context.Background(), cmd.Containerd.Namespace)
		)
	
		client, err := containerd.New(address)
		if err != nil {
			return err
		}
		defer client.Close()

		image, err := client.Pull(
			ctx,
			"docker.io/library/redis:alpine",
			containerd.WithPullUnpack,
		)
		if err != nil {
			return err
		}
	
		container, err := client.NewContainer(
			ctx,
			"redis",
			containerd.WithImage(image),
			containerd.WithNewSnapshot("redis-snapshot", image),
			containerd.WithNewSpec(oci.WithImageConfig(image)),
		)
		if err != nil {
			return err
		}
		defer container.Delete(ctx, containerd.WithSnapshotCleanup)

		task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
		if err != nil {
			return err
		}
		defer task.Delete(ctx)

		if _, err = task.Wait(ctx); err != nil {
			return err
		}
		if err = task.Start(ctx); err != nil {
			return err
		}

		close(ready)
		<-signals
		return nil
	}), nil
}
