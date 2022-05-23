package main

import "github.com/spf13/cobra"

var (
	getJobsCmd = &cobra.Command{
		Use:     "jobs",
		Aliases: []string{"job", "j"},
		RunE:    runGetJobs,
		Args:    cobra.RangeArgs(0, 1),
	}
)

func runGetJobs(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	var j any

	if len(args) > 0 {
		j, err = client.GetJob(args[0])
	} else {
		j, err = client.GetJobs()
	}
	if err != nil {
		return err
	}

	return write(j)
}
