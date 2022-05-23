package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	getStorageCmd = &cobra.Command{
		Use:     "storage",
		Aliases: []string{"storages", "s"},
		RunE:    runGetStorage,
		Args:    cobra.RangeArgs(0, 1),
	}
	createStorageCmd = &cobra.Command{
		Use:     "storage",
		Aliases: []string{"s"},
		RunE:    runCreateStorage,
		Args:    cobra.RangeArgs(0, 1),
	}
)

func init() {
	flags := getStorageCmd.PersistentFlags()
	flags.StringP("input-of", "i", "", "Input of job")
	_ = viper.BindPFlag("input-of", flags.Lookup("input-of"))
	flags.StringP("output-of", "o", "", "Output of job")
	_ = viper.BindPFlag("output-of", flags.Lookup("output-of"))
}

func runGetStorage(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	var (
		s any
		i = viper.GetString("input-of")
		o = viper.GetString("output-of")
	)

	switch {
	case len(args) > 0:
		s, err = client.GetStorage(args[0])
	case o != "":
		s, err = client.GetJobOutput(o)
	case i != "":
		s, err = client.GetJobInput(i)
	default:
		s, err = client.GetStorages()
	}
	if err != nil {
		return err
	}

	return write(s)
}

func runCreateStorage(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	i, err := getInput()
	if err != nil {
		return err
	}

	s, err := client.CreateStorage(i, viper.GetString("name"))
	if err != nil {
		return err
	}

	return write(s)
}
