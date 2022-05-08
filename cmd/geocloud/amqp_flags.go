package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/logsquaredn/geocloud/messagequeue"
)

var (
	amqpAddress string
	amqpOpts    = &messagequeue.AMQPMessageQueueOpts{}
)

func init() {
	rootCmd.PersistentFlags().StringVar(
		&amqpAddress,
		"amqp-address",
		coalesceString(
			os.Getenv("GEOCLOUD_AMQP_ADDRESS"),
			fmt.Sprintf(":%d", defaultAMQPPort),
		),
		"AMQP address",
	)
	rootCmd.PersistentFlags().StringVar(
		&amqpOpts.User,
		"amqp-user",
		coalesceString(
			os.Getenv("GEOCLOUD_AMQP_USER"),
			"geocloud",
		),
		"AMQP user",
	)
	rootCmd.PersistentFlags().StringVar(
		&amqpOpts.Password,
		"amqp-password",
		os.Getenv("GEOCLOUD_AMQP_PASSWORD"),
		"AMQP password",
	)
	rootCmd.PersistentFlags().Int64Var(
		&amqpOpts.Retries,
		"amqp-retries",
		parseInt64(os.Getenv("GEOCLOUD_AMQP_RETRIES")),
		"AMQP retries",
	)
	rootCmd.PersistentFlags().DurationVar(
		&amqpOpts.RetryDelay,
		"amqp-retry-delay",
		s5,
		"AMQP retry delay",
	)
	rootCmd.PersistentFlags().StringVar(
		&amqpOpts.QueueName,
		"amqp-queue-name",
		coalesceString(
			os.Getenv("GEOCLOUD_AMQP_QUEUE_NAME"),
			"geocloud",
		),
		"AMQP queue name",
	)
}

func getAMQPOpts() *messagequeue.AMQPMessageQueueOpts {
	delimiter := strings.Index(amqpAddress, ":")
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
