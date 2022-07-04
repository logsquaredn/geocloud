package main

import (
	"fmt"
	"time"
)

const (
	localhost           = "localhost"
	defaultPostgresHost = localhost
	defaultPostgresPort = 5432
	defaultAMQPHost     = localhost
	defaultAMQPPort     = 5672
	defaultUser         = "rototiller"
	defaultAMQPUser     = defaultUser
	defaultPostgresUser = defaultUser
)

var (
	defaultPostgresAddress = fmt.Sprintf("%s:%d", defaultPostgresHost, defaultPostgresPort)
	defaultAMQPAddress     = fmt.Sprintf("%s:%d", defaultAMQPHost, defaultAMQPPort)
)

var defaultAMQPQueueName = "rototiller"

var (
	s5  time.Duration
	s4  time.Duration
	h24 time.Duration
)

func init() {
	s5, _ = time.ParseDuration("5s")
	s4, _ = time.ParseDuration("4s")
	h24, _ = time.ParseDuration("24h")
}
