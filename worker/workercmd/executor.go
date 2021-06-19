package workercmd

import (
	"context"
	"net/http"

	"github.com/containerd/containerd/remotes/docker"
	"github.com/logsquaredn/geocloud/worker"
)

func (cmd *WorkerCmd) executor(ctx context.Context, name string) (*worker.ExecutorRunner, error) {
	resolver := docker.NewResolver(docker.ResolverOptions{
		Client: http.DefaultClient,
	})
	if cmd.RegistryUsername != "" || cmd.RegistryPassword != "" {
		resolver = docker.NewResolver(docker.ResolverOptions{
			Authorizer: docker.NewDockerAuthorizer(
				docker.WithAuthCreds(func(s string) (string, string, error) {
					return cmd.RegistryUsername, cmd.RegistryPassword, nil
				}),
			),
		})
	}

	return worker.NewExecutorRunner(ctx, name, resolver, cmd.Registry)
}
