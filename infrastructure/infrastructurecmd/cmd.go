package infrastructurecmd

import "fmt"

type InfrastructureCmd struct {}

func (cmd *InfrastructureCmd) Execute(args []string) error {
	return fmt.Errorf("infrastructure not implemented")
}
