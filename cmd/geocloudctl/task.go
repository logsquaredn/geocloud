package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	getTasksCmd = &cobra.Command{
		Use:     "tasks",
		Aliases: []string{"task", "t"},
		RunE:    runGetTasks,
		Args:    cobra.RangeArgs(0, 1),
	}
)

func init() {
	flags := getTasksCmd.PersistentFlags()
	flags.StringP("job", "j", "", "Job")
	_ = viper.BindPFlag("job", flags.Lookup("job"))
}

func runGetTasks(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	var (
		t any
		j = viper.GetString("job")
	)

	switch {
	case len(args) > 0:
		t, err = client.GetTask(args[0])
	case j != "":
		t, err = client.GetJobTask(j)
	default:
		t, err = client.GetTasks()
	}
	if err != nil {
		return err
	}

	return write(t)
}
