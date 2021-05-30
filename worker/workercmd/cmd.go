package workercmd

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
)

type WorkerCmd struct {
	Containerd Containerd     `group:"Containerd" namespace:"containerd"`
	QueueName  *string	 	  `long:"queue-name" short:"q" description:"SQS queue name to listen on"`
	Tasks      []string       `long:"task" short:"t" description:"Tasks that the worker can run"`
}

type Containerd struct {
	NoRun    bool           `long:"no-run" description:"Whether or not to run its own containerd process. If specified, must target an already-running containerd address"`
	Bin      string         `long:"bin" default:"containerd" description:"Path to a containerd executable"`
	LogLevel string         `long:"log-level" default:"debug" choice:"trace" choice:"debug" choice:"info" choice:"warn" choice:"error" choice:"fatal" choice:"panic" description:"Containerd log level"`
	Namespace string        `long:"namespace" default:"geocloud" description:"Containerd namespace"`
	Address  flags.Filename `long:"address" default:"/run/geocloud/containerd/containerd.sock" description:"Address for containerd's GRPC server'"`
	Root     flags.Filename `long:"root" default:"/var/lib/geocloud/containerd" description:"Containerd root directory"`
	Config   flags.Filename `long:"config" default:"/etc/geocloud/containerd/config.toml" description:"Path to config file"`
	State    flags.Filename `long:"state" default:"/run/geocloud/containerd" description:"Containerd state directory"`
}

func (cmd *WorkerCmd) Execute(args []string) error {
	worker, err := cmd.worker()
	if err != nil {
		return err
	}

	members := grouper.Members{
		{
			Name: "worker",
			Runner: worker,
		},
	}

	if !cmd.Containerd.NoRun {
		containerd, err := cmd.containerd()
		if err != nil {
			return err
		}

		members = append(grouper.Members{
			{
				Name: "containerd",
				Runner: containerd,
			},
		}, members...)
	}

	return <-ifrit.Invoke(grouper.NewOrdered(os.Interrupt, members),).Wait()
}
