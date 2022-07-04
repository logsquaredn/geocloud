package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/logsquaredn/rototiller"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "rototiller",
	Version:           rototiller.Semver(),
	PersistentPreRunE: persistentPreRun,
}

var (
	loglevel string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&loglevel, "loglevel", "l", "info", "Loglevel")
	rootCmd.SetVersionTemplate(fmt.Sprintf("{{ .Name }}{{ .Version }} %s\n", runtime.Version()))
	rootCmd.AddCommand(
		migrateCmd,
		apiCmd,
		workerCmd,
		secretaryCmd,
	)
}

func persistentPreRun(cmd *cobra.Command, args []string) error {
	loglevel, err := zerolog.ParseLevel(loglevel)
	if err == nil {
		zerolog.SetGlobalLevel(loglevel)
	}
	return err
}

func main() {
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
