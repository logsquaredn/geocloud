package main

import (
	"strings"

	"github.com/logsquaredn/geocloud/messagequeue"
	"github.com/spf13/viper"
)

func init() {
	bindConfToFlags(rootCmd.PersistentFlags(), []*conf{
		{
			arg:  "amqp-address",
			def:  defaultAMQPAddress,
			desc: "AMQP address",
		},
		{
			arg:  "amqp-user",
			def:  defaultAMQPUser,
			desc: "AMQP user",
		},
		{
			arg:  "amqp-password",
			def:  "",
			desc: "AMQP password",
		},
		{
			arg:  "amqp-retries",
			def:  int64(5),
			desc: "AMQP retries",
		},
		{
			arg:  "amqp-retry-delay",
			def:  s5,
			desc: "AMQP retry delay",
		},
		{
			arg:  "amqp-queue-name",
			def:  defaultAMQPQueueName,
			desc: "AMQP queue name",
		},
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
		amqpOpts.Port = parseInt64(amqpAddress[delimiter:])
	}

	if amqpOpts.Host == "" {
		amqpOpts.Host = defaultAMQPHost
	}

	if amqpOpts.Port == 0 {
		amqpOpts.Port = defaultAMQPPort
	}

	return amqpOpts
}
