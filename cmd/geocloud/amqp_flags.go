package main

import (
	"strconv"
	"strings"

	"github.com/logsquaredn/geocloud/messagequeue"
	"github.com/spf13/viper"
)

func init() {
	bindConfToFlags(rootCmd.PersistentFlags(), []*conf{
		{"amqp-address", defaultAMQPAddress, "AMQP address"},
		{"amqp-user", defaultAMQPUser, "AMQP user"},
		{"amqp-password", "", "AMQP password"},
		{"amqp-retries", int64(5), "AMQP retries"},
		{"amqp-retry-delay", s5, "AMQP retry delay"},
		{"amqp-queue-name", defaultAMQPQueueName, "AMQP queue name"},
	}...)
}

func getAMQPOpts() *messagequeue.AMQPOpts {
	var (
		amqpOpts = &messagequeue.AMQPOpts{
			User:       viper.GetString("amqp-user"),
			Password:   viper.GetString("amqp-password"),
			Retries:    viper.GetInt64("amqp-retries"),
			RetryDelay: viper.GetDuration("amqp-retry-delay"),
			QueueName:  viper.GetString("amqp-queue-name"),
		}
		amqpAddress = viper.GetString("amqp-address")
		delimiter   = strings.Index(amqpAddress, ":")
	)
	if delimiter < 0 {
		amqpOpts.Host = amqpAddress
	} else {
		amqpOpts.Host = amqpAddress[:delimiter]
		port, _ := strconv.Atoi(amqpAddress[delimiter:])
		amqpOpts.Port = int64(port)
	}

	if amqpOpts.Host == "" {
		amqpOpts.Host = defaultAMQPHost
	}

	if amqpOpts.Port == 0 {
		amqpOpts.Port = defaultAMQPPort
	}

	return amqpOpts
}
