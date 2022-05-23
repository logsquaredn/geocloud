package main

import (
	"context"
	"os"

	"github.com/logsquaredn/geocloud"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:     "geocloudctl",
		Version: geocloud.Semver(),
	}
	getCmd = &cobra.Command{
		Use: "get",
	}
)

func init() {
	flags := rootCmd.PersistentFlags()
	flags.String("api-key", "", "Geocloud API key")
	viper.BindPFlag("api-key", flags.Lookup("api-key"))
	flags.String("base-url", "", "Geocloud base URL")
	viper.BindPFlag("base-url", flags.Lookup("base-url"))
}

func init() {
	getCmd.AddCommand(
		getJobsCmd,
		getStorageCmd,
		getTasksCmd,
	)
	rootCmd.AddCommand(
		getCmd,
	)
}

func main() {
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
