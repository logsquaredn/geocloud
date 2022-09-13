package command

import (
	"encoding/json"
	"os"

	"github.com/logsquaredn/rototiller/pkg/client"
	"github.com/spf13/cobra"
)

func NewGetTasksCommand() *cobra.Command {
	var (
		addr, apiKey string
		getTasksCmd  = &cobra.Command{
			Use:     "tasks",
			Aliases: []string{"task, t"},
			Args:    cobra.MaximumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				c, err := client.New(addr, apiKey)
				if err != nil {
					cmd.PrintErrln(err)
					os.Exit(1)
				}

				var a any
				if len(args) > 0 {
					a, err = c.GetTask(args[0])
				} else {
					a, err = c.GetTasks()
				}
				if err != nil {
					cmd.PrintErrln(err)
					os.Exit(1)
				}

				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")

				if err = encoder.Encode(a); err != nil {
					cmd.PrintErrln(err)
					os.Exit(1)
				}
			},
		}
	)

	getTasksCmd.Flags().StringVar(&addr, "addr", "", "rototiller address")
	getTasksCmd.Flags().StringVar(&apiKey, "api-key", "", "rototiller API key")

	return getTasksCmd
}
