package mockcmd

import (
	"fmt"
	"os"
)

// MockCmd implements a command. It copies an input file to an output destination
type MockCmd struct {}

// Execute copies an input file to an output destination
func (cmd *MockCmd) Execute(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("not enough args")
	}

	i, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}

	err = os.WriteFile(args[1], i, 0644)
	if err != nil {
		return err
	}

	return nil
}
