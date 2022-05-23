package main

import (
	"strconv"
	"strings"

	"github.com/logsquaredn/geocloud/datastore"
	"github.com/spf13/viper"
)

func init() {
	bindConfToFlags(rootCmd.PersistentFlags(), []*conf{
		{"postgres-address", defaultPostgresAddress, "Postgres address"},
		{"postgres-user", defaultPostgresUser, "Postgres user"},
		{"postgres-password", "", "Postgres password"},
		{"postgres-retries", int64(5), "Postgres retries"},
		{"postgres-retry-delay", s5, "Postgres retry delay"},
		{"postgres-sslmode", "", "Postgres SSL mode"},
	}...)
}

func getPostgresOpts() *datastore.PostgresOpts {
	var (
		postgresOpts = &datastore.PostgresOpts{
			User:       viper.GetString("postgres-user"),
			Password:   viper.GetString("postgres-password"),
			Retries:    viper.GetInt64("postgres-retries"),
			RetryDelay: viper.GetDuration("postgres-retry-delay"),
			SSLMode:    viper.GetString("postgres-sslmode"),
		}
		postgresAddress = viper.GetString("postgres-address")
		delimiter       = strings.Index(postgresAddress, ":")
	)
	if delimiter < 0 {
		postgresOpts.Host = postgresAddress
	} else {
		postgresOpts.Host = postgresAddress[:delimiter]
		port, _ := strconv.Atoi(postgresAddress[delimiter:])
		postgresOpts.Port = int64(port)
	}

	if postgresOpts.Host == "" {
		postgresOpts.Host = defaultPostgresHost
	}

	if postgresOpts.Port == 0 {
		postgresOpts.Port = defaultPostgresPort
	}

	return postgresOpts
}
