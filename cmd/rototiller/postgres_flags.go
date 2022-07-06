package main

import (
	"strconv"
	"strings"

	"github.com/logsquaredn/rototiller/datastore"
	"github.com/logsquaredn/rototiller/internal/conf"
	"github.com/spf13/viper"
)

func init() {
	_ = conf.BindToFlags(rootCmd.PersistentFlags(), nil, []*conf.Conf{
		{
			Arg:         "postgres-address",
			Default:     defaultPostgresAddress,
			Description: "Postgres address",
		},
		{
			Arg:         "postgres-user",
			Default:     defaultPostgresUser,
			Description: "Postgres user",
		},
		{
			Arg:         "postgres-password",
			Default:     "",
			Description: "Postgres password",
		},
		{
			Arg:         "postgres-retries",
			Default:     int64(5),
			Description: "Postgres retries",
		},
		{
			Arg:         "postgres-retry-delay",
			Default:     s5,
			Description: "Postgres retry delay",
		},
		{
			Arg:         "postgres-sslmode",
			Default:     "",
			Description: "Postgres SSL mode",
		},
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
