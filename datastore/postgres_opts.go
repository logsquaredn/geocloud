package datastore

import (
	"fmt"
	"time"
)

type PostgresDatastoreOpts struct {
	Host       string
	Port       int64
	User       string
	Password   string
	SSLMode    string
	Retries    int64
	RetryDelay time.Duration
}

func (o *PostgresDatastoreOpts) connectionString() string {
	sslmode := "disable"
	if o.SSLMode != "" {
		sslmode = o.SSLMode
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d?sslmode=%s", o.User, o.Password, o.Host, o.Port, sslmode)
}
