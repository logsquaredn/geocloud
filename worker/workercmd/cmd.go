package workercmd

import (
	"fmt"

	"github.com/jessevdk/go-flags"
)

type WorkerCmd struct {
	WorkDir flags.Filename `long:"work-dir" default:"/var/geocloud/worker" description:"directory in which to store worker data"`
	Tasks   []string       `long:"task" short:"t" description:"tasks that the worker can run"`
}

func (cmd *WorkerCmd) Execute(args []string) error {
	return fmt.Errorf("worker not implemented")
}
