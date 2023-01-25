package command

import (
	"encoding/json"

	"github.com/logsquaredn/rototiller/client"
	"github.com/spf13/cobra"
)

func NewGetTasks() *cobra.Command {
	var (
		addr, apiKey string
		cmd          = &cobra.Command{
			Use:     "tasks",
			Aliases: []string{"task, t"},
			Args:    cobra.MaximumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				c, err := client.New(addr, apiKey)
				if err != nil {
					return err
				}

				var a any
				if len(args) > 0 {
					a, err = c.GetTask(args[0])
				} else {
					a, err = c.GetTasks()
				}
				if err != nil {
					return err
				}

				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")

				return encoder.Encode(a)
			},
		}
	)

	cmd.Flags().StringVar(&addr, "addr", "", "rototiller address")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "rototiller API key")

	return cmd
}
