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
			Use:     "rotoctl",
			Version: rototiller.GetSemver(),
			PersistentPreRun: func(cmd *cobra.Command, args []string) {
				cmd.SetContext(rototiller.WithLogger(cmd.Context(), rototiller.NewLogger().V(verbosity)))
			},
		}
		createCmd = &cobra.Command{
			Use:     "create",
			Aliases: []string{"c"},
		}
		getCmd = &cobra.Command{
			Use:     "get",
			Aliases: []string{"g"},
		}
		runCmd = &cobra.Command{
			Use:     "run",
			Aliases: []string{"r"},
		}
	)

	createCmd.AddCommand(NewCreateJob())
	getCmd.AddCommand(NewGetJob(), NewGetTasks())
	runCmd.AddCommand(NewRunJob())

	cmd.PersistentFlags().CountVarP(&verbosity, "verbose", "V", "verbose")
	cmd.SetVersionTemplate("{{ .Name }}{{ .Version }} " + runtime.Version() + "\n")
	cmd.AddCommand(createCmd, getCmd, runCmd)

	return cmd
}
