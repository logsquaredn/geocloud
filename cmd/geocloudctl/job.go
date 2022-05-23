package main

import (
	"github.com/logsquaredn/geocloud"
	"github.com/spf13/cobra"
)

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

func init() {
	flags := createJobCmd.PersistentFlags()
	flags.String("input", "", "Storage ID to use")
	flags.String("input-of", "", "Job ID to use the input of")
	flags.String("output-of", "", "Job ID to use the output of")
}

func init() {
	flags := runJobCmd.PersistentFlags()
	flags.String("input", "", "Storage ID to use")
	flags.String("input-of", "", "Job ID to use the input of")
	flags.String("output-of", "", "Job ID to use the output of")
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

func runCreateJob(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	var (
		req geocloud.Request
		i   = cmd.Flag("input").Value.String()
		io  = cmd.Flag("input-of").Value.String()
		oo  = cmd.Flag("output-of").Value.String()
	)
	switch {
	case i != "":
		req = geocloud.NewJobWithInput(i)
	case io != "":
		req = geocloud.NewJobWithInputOfJob(io)
	case oo != "":
		req = geocloud.NewJobWithOutputOfJob(oo)
	default:
		f, err := getInput(cmd)
		if err != nil {
			return err
		}
		req = geocloud.NewJobFromInput(f)
	}

	j, err := client.CreateJob(args[0], req)
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

	var (
		req geocloud.Request
		i   = cmd.Flag("input").Value.String()
		io  = cmd.Flag("input-of").Value.String()
		oo  = cmd.Flag("output-of").Value.String()
	)
	switch {
	case i != "":
		req = geocloud.NewJobWithInput(i)
	case io != "":
		req = geocloud.NewJobWithInputOfJob(io)
	case oo != "":
		req = geocloud.NewJobWithOutputOfJob(oo)
	default:
		f, err := getInput(cmd)
		if err != nil {
			return err
		}
		req = geocloud.NewJobFromInput(f)
	}

	j, err := client.RunJob(args[0], req)
	if err != nil {
		return err
	}

	return write(j)
}
