package main

import "github.com/spf13/cobra"

var (
	getStorageCmd = &cobra.Command{
		Use:  "storage",
		Aliases: []string{"storages", "s"},
		RunE: runGetStorage,
		Args: cobra.RangeArgs(0, 1),
	}
)

func runGetStorage(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	var s any

	if len(args) > 0 {
		s, err = client.GetStorage(args[0])
	} else {
		s, err = client.GetStorages()
	}
	if err != nil {
		return err
	}

	return write(s)
}
