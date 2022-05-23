package main

import (
	"strings"

	"github.com/logsquaredn/geocloud/datastore"
	"github.com/spf13/viper"
)

var (
	postgresAddress string
	postgresOpts    = &datastore.PostgresOpts{}
)

func init() {
	bindConfToFlags(rootCmd.PersistentFlags(), []*conf{
		{
			arg:  "postgres-address",
			def:  defaultPostgresAddress,
			desc: "Postgres address",
		},
		{
			arg:  "postgres-user",
			def:  defaultPostgresUser,
			desc: "Postgres user",
		},
		{
			arg:  "postgres-password",
			def:  "",
			desc: "Postgres password",
		},
		{
			arg:  "postgres-retries",
			def:  int64(5),
			desc: "Postgres retries",
		},
		{
			arg:  "postgres-retry-delay",
			def:  s5,
			desc: "Postgres retry delay",
		},
		{
			arg:  "postgres-sslmode",
			def:  "",
			desc: "Postgres SSL mode",
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
		postgresOpts.Port = parseInt64(postgresAddress[delimiter:])
	}

	if postgresOpts.Host == "" {
		postgresOpts.Host = defaultPostgresHost
	}

	if postgresOpts.Port == 0 {
		postgresOpts.Port = defaultPostgresPort
	}

	return postgresOpts
}
