package mockcmd

import (
	"os"

	"github.com/jessevdk/go-flags"
)

// MockCmd implements a command. It copies an input file to an output destination
type MockCmd struct {
	Input  flags.Filename `long:"input" short:"i" required:"true" description:"path to input file"`
	Output flags.Filename `long:"output" short:"o" required:"true" description:"path to output file"`
}

// Execute copies an input file to an output destination
func (cmd *MockCmd) Execute(args []string) error {
	i, err := os.ReadFile(string(cmd.Input))
	if err != nil {
		return err
	}

	err = os.WriteFile(string(cmd.Output), i, 0644)
	if err != nil {
		return err
	}

	return nil
}
