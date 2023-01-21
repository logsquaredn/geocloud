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

	createCmd.AddCommand(NewCreateJobCommand())
	getCmd.AddCommand(NewGetJobCommand(), NewGetTasksCommand())
	runCmd.AddCommand(NewRunJobCommand())

	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "V", "verbose")
	rootCmd.SetVersionTemplate("{{ .Name }}{{ .Version }} " + runtime.Version() + "\n")
	rootCmd.AddCommand(createCmd, getCmd, runCmd)

	return rootCmd
}
