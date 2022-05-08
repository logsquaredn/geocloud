package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/logsquaredn/geocloud/datastore"
)

var (
	postgresAddress string
	postgresOpts    = &datastore.PostgresDatastoreOpts{}
)

func init() {
	rootCmd.PersistentFlags().StringVar(
		&postgresAddress,
		"postgres-address",
		coalesceString(
			os.Getenv("GEOCLOUD_POSTGRES_ADDRESS"),
			fmt.Sprintf(":%d", defaultPostgresPort),
		),
		"Postgres address",
	)
	rootCmd.PersistentFlags().StringVar(
		&postgresOpts.User,
		"postgres-user",
		coalesceString(
			os.Getenv("GEOCLOUD_POSTGRES_USER"),
			"geocloud",
		),
		"Postgres user",
	)
	rootCmd.PersistentFlags().StringVar(
		&postgresOpts.Password,
		"postgres-password",
		os.Getenv("GEOCLOUD_POSTGRES_PASSWORD"),
		"Postgres password",
	)
	rootCmd.PersistentFlags().StringVar(
		&postgresOpts.SSLMode,
		"postgres-sslmode",
		os.Getenv("GEOCLOUD_POSTGRES_SSLMODE"),
		"Postgres SSL mode",
	)
	rootCmd.PersistentFlags().Int64Var(
		&postgresOpts.Retries,
		"postgres-retries",
		parseInt64(os.Getenv("GEOCLOUD_POSTGRES_RETRIES")),
		"Postgres retries",
	)
	rootCmd.PersistentFlags().DurationVar(
		&postgresOpts.RetryDelay,
		"postgres-retry-delay",
		s5,
		"Postgres retry delay",
	)
}

func getPostgresOpts() *datastore.PostgresDatastoreOpts {
	delimiter := strings.Index(postgresAddress, ":")
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
