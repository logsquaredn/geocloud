package worker

import (
	"context"
	"fmt"
	"os"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/remotes"
	"github.com/logsquaredn/geocloud"
)

type ExecutorRunner struct {
	clientChan    chan *containerd.Client
	ctx           context.Context
	name          string
	registry      *Registry
	messageChan   chan *geocloud.Message
	resolver   	  remotes.Resolver
}

func (r *ExecutorRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	var (
		ctx = r.ctx
		client = <-r.clientChan
		messages = r.messageChan
		name = r.name
		registry = r.registry
		resolver = r.resolver
	)
	if client == nil {
		return fmt.Errorf("nil client supplied to executor")
	}

	wait := make(chan error, 1)
	close(ready)
	for i := 1;; i++ {
		select {
		case message := <-messages:
			id := fmt.Sprintf("%s_%d", name, i)
			image, err := client.Pull(ctx, registry.Ref(message.Image), containerd.WithPullUnpack, containerd.WithResolver(resolver))
			if err != nil {
				wait<- err
			}

			container, err := client.NewContainer(
				ctx,
				id,
				containerd.WithImage(image),
				containerd.WithNewSnapshot(id, image),
				containerd.WithNewSpec(oci.WithImageConfig(image)),
			)
			if err != nil {
				wait<- err
			}
	
			task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
			if err != nil {
				wait<- err
			}
	
			if _, err = task.Wait(ctx); err != nil {
				wait<- err
			}
			if err = task.Start(ctx); err != nil {
				wait<- err
			}

			task.Delete(ctx)
			container.Delete(ctx, containerd.WithSnapshotCleanup)
		case <-signals:
			return nil
		case err := <-wait:
			return err
		}
	}
}

func (r *ExecutorRunner) ClientChannel() chan *containerd.Client {
	return r.clientChan
}

func (r *ExecutorRunner) MessageChannel() chan *geocloud.Message {
	return r.messageChan
}

func (r *ExecutorRunner) Name() string {
	return r.name
}

func NewExecutorRunner(ctx context.Context, name string, resolver remotes.Resolver, registryurl string) (*ExecutorRunner, error) {
	registry, err := NewRegistry(registryurl)
	if err != nil {
		return nil, err
	}

	return &ExecutorRunner{
		clientChan: make(chan *containerd.Client, 1),
		ctx: ctx,
		messageChan: make(chan *geocloud.Message),
		name: name,
		registry: registry,
		resolver: resolver,
	}, nil
}
