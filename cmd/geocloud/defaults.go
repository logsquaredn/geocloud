package main

import "time"

const (
	localhost           = "localhost"
	defaultPostgresHost = localhost
	defaultPostgresPort = 5432
	defaultAMQPHost     = localhost
	defaultAMQPPort     = 5672
)

var (
	s5  time.Duration
	h24 time.Duration
)

func init() {
	s5, _ = time.ParseDuration("5s")
	h24, _ = time.ParseDuration("24h")
}
