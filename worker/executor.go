package worker

import (
	"context"
	"fmt"
	"os"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/oci"
)

type ExecutorRunner struct {
	clientChannel chan *containerd.Client
	ctx           context.Context
	name          string
}

func (r *ExecutorRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	ctx := r.ctx
	client := <-r.clientChannel
	if client == nil {
		return fmt.Errorf("nil client supplied to executor")
	}
	
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
}

func (r *ExecutorRunner) ClientChannel() chan *containerd.Client {
	return r.clientChannel
}

func (r *ExecutorRunner) Name() string {
	return r.name
}

func NewExecutorRunner(ctx context.Context, name string) (*ExecutorRunner, error) {
	return &ExecutorRunner{
		clientChannel: make(chan *containerd.Client),
		ctx: ctx,
		name: name,
	}, nil
}
