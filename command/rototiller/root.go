package command

import (
	"runtime"

	"github.com/logsquaredn/rototiller"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	var (
		verbosity int
		cmd       = &cobra.Command{
			Use:     "rototiller",
			Version: rototiller.GetSemver(),
			PersistentPreRun: func(cmd *cobra.Command, args []string) {
				cmd.SetContext(rototiller.WithLogger(cmd.Context(), rototiller.NewLogger().V(verbosity)))
			},
		}
	)

	cmd.PersistentFlags().CountVarP(&verbosity, "verbose", "V", "verbose")
	cmd.SetVersionTemplate("{{ .Name }}{{ .Version }} " + runtime.Version() + "\n")
	cmd.AddCommand(NewAPI(), NewWorker(), NewMigrate(), NewSecretary())

	return cmd
}
