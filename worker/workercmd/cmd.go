package workercmd

import (
	"context"
	"fmt"
	"os"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/jessevdk/go-flags"
	"github.com/logsquaredn/geocloud/worker"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
)

type WorkerCmd struct {
	Containerd Containerd `group:"Containerd" namespace:"containerd"`
	Executors  int 		  `long:"executors" short:"e" default:"1" description:"Number of threads to run tasks on"`
	Loglevel   string     `long:"log-level" short:"l" description:"Geocloud log level"`
	QueueNames []*string  `long:"queue-name" short:"q" description:"SQS queue names to listen on"`
}

type Containerd struct {
	NoRun     bool           `long:"no-run" description:"Whether or not to run its own containerd process. If specified, must target an already-running containerd address"`
	Bin       string         `long:"bin" default:"containerd" description:"Path to a containerd executable"`
	Address   flags.Filename `long:"address" default:"/run/geocloud/containerd/containerd.sock" description:"Address for containerd's GRPC server'"`
	Root      flags.Filename `long:"root" default:"/var/lib/geocloud/containerd" description:"Containerd root directory"`
	State     flags.Filename `long:"state" default:"/run/geocloud/containerd" description:"Containerd state directory"`
	Config    flags.Filename `long:"config" default:"/etc/geocloud/containerd/config.toml" description:"Path to config file"`
	Loglevel  string         `long:"log-level" default:"debug" choice:"trace" choice:"debug" choice:"info" choice:"warn" choice:"error" choice:"fatal" choice:"panic" description:"Containerd log level"`
	Namespace string         `long:"namespace" default:"geocloud" description:"Containerd namespace"`
}

func (cmd *WorkerCmd) Execute(args []string) error {
	var members grouper.Members

	if !cmd.Containerd.NoRun {
		containerd, err := cmd.containerd()
		if err != nil {
			return err
		}

		members = append(members, grouper.Member{
			Name: "containerd",
			Runner: containerd,
		})
	}

	ctx := namespaces.WithNamespace(context.Background(), cmd.Containerd.Namespace)
	executors := make([]*worker.ExecutorRunner, cmd.Executors)
	for i := 0; i < cmd.Executors; i++ {
		if executor, err := cmd.executor(ctx, fmt.Sprintf("executor_%d", i)); err == nil {
			executors[i] = executor
		} else {
			return err
		}
	}

	members = append(members, grouper.Member{
		Name: "injector",
		Runner: ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
			client, err := containerd.New(string(cmd.Containerd.Address))
			if err != nil {
				return err
			}

			go func() {
				for _, executor := range executors {
					channel := executor.ClientChannel()
					channel<- client
					close(channel)
				}
			}()

			close(ready)
			<-signals
			return nil
		}),
	})

	for _, executor := range executors {
		members = append(members, grouper.Member{
			Name: executor.Name(),
			Runner: executor,
		})
	}

	return <-ifrit.Invoke(grouper.NewOrdered(os.Interrupt, members)).Wait()
}
