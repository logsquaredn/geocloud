package apicmd

import (
	"fmt"
)

type APICmd struct {}

func (cmd *APICmd) Execute(args []string) error {
	return fmt.Errorf("api not implemented")
}
