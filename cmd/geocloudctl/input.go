package main

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

func getInput(cmd *cobra.Command) (io.Reader, error) {
	return os.Open(cmd.Flag("file").Value.String())
}
