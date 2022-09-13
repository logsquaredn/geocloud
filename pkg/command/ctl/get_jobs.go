package command

import (
	"encoding/json"
	"os"

	"github.com/logsquaredn/rototiller/pkg/client"
	"github.com/spf13/cobra"
)

func NewGetJobCommand() *cobra.Command {
	var (
		addr, apiKey, contentType string
		query                     = map[string]string{}
		rpc                       bool
		getJobCmd                 = &cobra.Command{
			Use:     "jobs",
			Aliases: []string{"job", "j"},
			Args:    cobra.MaximumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				var (
					a    any
					opts = []client.ClientOpt{}
				)
				if rpc {
					opts = append(opts, client.WithRPC)
				}

				c, err := client.New(addr, apiKey, opts...)
				if err != nil {
					cmd.PrintErrln(err)
					os.Exit(1)
				}

				if len(args) > 0 {
					a, err = c.GetJob(args[0])
				} else {
					a, err = c.GetJobs()
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

	getJobCmd.Flags().BoolVar(&rpc, "rpc", false, "use RPC")
	getJobCmd.Flags().StringVar(&addr, "addr", "", "rototiller address")
	getJobCmd.Flags().StringVar(&apiKey, "api-key", "", "rototiller API key")
	getJobCmd.Flags().StringVar(&contentType, "content-type", "", "content-type")
	getJobCmd.Flags().StringToStringVarP(&query, "query", "q", map[string]string{}, "query params")

	return getJobCmd
}
