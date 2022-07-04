package main

import (
	"fmt"
	"os/user"

	"github.com/logsquaredn/rototiller"
	"github.com/spf13/viper"
)

func getClient() (*rototiller.Client, error) {
	port := viper.GetInt64("port")
	if port == 0 {
		port = 8080
	}

	var (
		baseURL = fmt.Sprintf("http://localhost:%d/", port)
		apiKey  = viper.GetString("api-key")
	)
	if c := viper.GetString("base-url"); c != "" {
		baseURL = c
	} else if apiKey == "" {
		// hack for local dev
		if u, err := user.Current(); err == nil {
			apiKey = u.Username
			if apiKey == "" {
				apiKey = u.Name
			}
		}
	}

	var (
		rpc  = viper.GetBool("rpc")
		opts = []rototiller.ClientOpt{}
	)
	if rpc {
		opts = append(opts, rototiller.WithRPC)
	}

	client, err := rototiller.NewClient(baseURL, apiKey, opts...)
	if err != nil {
		return nil, err
	}

	return client, nil
}
