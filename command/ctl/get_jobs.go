package command

import (
	"encoding/json"

	"github.com/logsquaredn/rototiller/client"
	"github.com/spf13/cobra"
)

func NewGetJob() *cobra.Command {
	var (
		addr, apiKey, contentType string
		query                     = map[string]string{}
		cmd                       = &cobra.Command{
			Use:     "jobs",
			Aliases: []string{"job", "j"},
			Args:    cobra.MaximumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				var (
					a    any
					opts = []client.ClientOpt{}
				)

				c, err := client.New(addr, apiKey, opts...)
				if err != nil {
					return err
				}

				if len(args) > 0 {
					a, err = c.GetJob(args[0])
				} else {
					a, err = c.GetJobs()
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
	cmd.Flags().StringVar(&contentType, "content-type", "", "content-type")
	cmd.Flags().StringToStringVarP(&query, "query", "q", map[string]string{}, "query params")

	return cmd
}
