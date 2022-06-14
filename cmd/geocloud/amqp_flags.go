package main

import (
	"strconv"
	"strings"

	"github.com/logsquaredn/geocloud/internal/conf"
	"github.com/logsquaredn/geocloud/messagequeue"
	"github.com/spf13/viper"
)

func init() {
	_ = conf.BindToFlags(rootCmd.PersistentFlags(), nil, []*conf.Conf{
		{
			Arg:         "amqp-address",
			Default:     defaultAMQPAddress,
			Description: "AMQP address",
		},
		{
			Arg:         "amqp-user",
			Default:     defaultAMQPUser,
			Description: "AMQP user",
		},
		{
			Arg:         "amqp-password",
			Default:     "",
			Description: "AMQP password",
		},
		{
			Arg:         "amqp-retries",
			Default:     int64(5),
			Description: "AMQP retries",
		},
		{
			Arg:         "amqp-retry-delay",
			Default:     s4,
			Description: "AMQP retry delay",
		},
		{
			Arg:         "amqp-queue-name",
			Default:     defaultAMQPQueueName,
			Description: "AMQP queue name",
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
