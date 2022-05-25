package main

import (
	"fmt"
	"os"
	"os/user"
	"strconv"

	"github.com/logsquaredn/geocloud"
	"github.com/spf13/viper"
)

func getClient() (*geocloud.Client, error) {
	port, _ := strconv.Atoi(os.Getenv("GEOCLOUD_PORT"))
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

	client, err := geocloud.NewClient(baseURL, apiKey)
	if err != nil {
		return nil, err
	}

	return client, nil
}
