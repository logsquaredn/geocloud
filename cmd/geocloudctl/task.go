package main

import "github.com/spf13/cobra"

var (
	getTasksCmd = &cobra.Command{
		Use:     "tasks",
		Aliases: []string{"task", "t"},
		RunE:    runGetTasks,
		Args:    cobra.RangeArgs(0, 1),
	}
)

func runGetTasks(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	var t any

	if len(args) > 0 {
		t, err = client.GetTask(args[0])
	} else {
		t, err = client.GetTasks()
	}
	if err != nil {
		return err
	}

	return write(t)
}
