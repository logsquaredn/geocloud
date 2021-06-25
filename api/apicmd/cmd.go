package apicmd

import (
	"fmt"
)

type APICmd struct {
	Version func() `short:"v" long:"version" description:"Print the version"`
}

func (cmd *APICmd) Execute(args []string) error {
	return fmt.Errorf("api not implemented")
}
