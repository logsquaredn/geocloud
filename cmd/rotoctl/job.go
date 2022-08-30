package main

import "github.com/spf13/cobra"

var (
	getJobsCmd = &cobra.Command{
		Use:     "jobs",
		Aliases: []string{"job", "j"},
		RunE:    runGetJobs,
		Args:    cobra.RangeArgs(0, 1),
	}
	runJobCmd = &cobra.Command{
		Use:     "job",
		Aliases: []string{"j"},
		RunE:    runRunJob,
		Args:    cobra.ExactArgs(1),
	}
	jobQuery = map[string]string{}
)

func init() {
	flags := runJobCmd.PersistentFlags()
	flags.String("input", "", "Storage ID to use")
	flags.String("input-of", "", "Job ID to use the input of")
	flags.String("output-of", "", "Job ID to use the output of")
	flags.String("content-type", "", "Content-Type to send. Auto detected if not supplied and 'application/json' if auto detect fails")
	flags.StringToStringVarP(&jobQuery, "query", "q", map[string]string{}, "Query params to send")
}

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

func runRunJob(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	req, err := getRequest(cmd)
	if err != nil {
		return err
	}

	j, err := client.RunJob(args[0], req)
	if err != nil {
		return err
	}

	return write(j)
}