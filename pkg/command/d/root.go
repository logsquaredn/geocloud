package command

import (
	"runtime"

	"github.com/logsquaredn/rototiller"
	"github.com/spf13/cobra"
)

func NewRoot() *cobra.Command {
	var (
		verbosity int
		rootCmd   = &cobra.Command{
			Use:     "rototiller",
			Version: rototiller.Semver(),
			PersistentPreRun: func(cmd *cobra.Command, args []string) {
				cmd.SetContext(rototiller.WithLogger(cmd.Context(), rototiller.NewLogger().V(verbosity)))
			},
		}
	)

	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "V", "verbose")
	rootCmd.SetVersionTemplate("{{ .Name }}{{ .Version }} " + runtime.Version() + "\n")
	rootCmd.AddCommand(NewAPI(), NewWorker(), NewMigrate(), NewSecretary(), NewPlumber())

	return rootCmd
}
