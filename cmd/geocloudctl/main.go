package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

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
	runCmd = &cobra.Command{
		Use: "run",
	}
	createCmd = &cobra.Command{
		Use: "create",
	}
)

func init() {
	flags := rootCmd.PersistentFlags()
	flags.String("api-key", "", "Geocloud API key")
	_ = viper.BindPFlag("api-key", flags.Lookup("api-key"))
	flags.String("base-url", "", "Geocloud base URL")
	_ = viper.BindPFlag("base-url", flags.Lookup("base-url"))
	flags.String("port", "", "Geocloud port")
	_ = viper.BindPFlag("port", flags.Lookup("port"))
	flags.BoolP("rpc", "r", false, "use RPC")
	_ = viper.BindPFlag("rpc", flags.Lookup("rpc"))
}

func init() {
	flags := createCmd.PersistentFlags()
	flags.StringP("file", "f", "", "Path to input file")
}

func init() {
	flags := runCmd.PersistentFlags()
	flags.StringP("file", "f", "", "Path to input file")
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("{{ .Name }}{{ .Version }} %s\n", runtime.Version()))
	getCmd.AddCommand(
		getJobsCmd,
		getStorageCmd,
		getTasksCmd,
	)
	runCmd.AddCommand(
		runJobCmd,
	)
	createCmd.AddCommand(
		createStorageCmd,
	)
	rootCmd.AddCommand(
		getCmd,
		runCmd,
		createCmd,
	)
}

func main() {
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
