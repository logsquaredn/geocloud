package main

import "github.com/spf13/cobra"

var (
	getJobsCmd = &cobra.Command{
		Use:     "jobs",
		Aliases: []string{"job", "j"},
		RunE:    runGetJobs,
		Args:    cobra.RangeArgs(0, 1),
	}
	createJobCmd = &cobra.Command{
		Use:     "job",
		Aliases: []string{"j"},
		RunE:    runCreateJob,
		Args:    cobra.ExactArgs(1),
	}
	runJobCmd = &cobra.Command{
		Use:     "job",
		Aliases: []string{"j"},
		RunE:    runRunJob,
		Args:    cobra.ExactArgs(1),
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

func runCreateJob(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	i, err := getInput()
	if err != nil {
		return err
	}

	j, err := client.CreateJob(args[0], i)
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

	i, err := getInput()
	if err != nil {
		return err
	}

	j, err := client.RunJob(args[0], i)
	if err != nil {
		return err
	}

	return write(j)
}
