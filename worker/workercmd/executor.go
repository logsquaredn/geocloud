package workercmd

import (
	"context"

	"github.com/logsquaredn/geocloud/worker"
)


func (cmd *WorkerCmd) executor(ctx context.Context, name string) (*worker.ExecutorRunner, error) {
	return worker.NewExecutorRunner(ctx, name)
}
