package router

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud/shared/das"
	"github.com/logsquaredn/geocloud/shared/oas"
	"github.com/rs/zerolog/log"
	"github.com/tedsuo/ifrit"
)

type f = map[string]interface{}

type Router struct {
	das   *das.Das
	oas   *oas.Oas
	sqs   *sqs.SQS
	group string
}

var _ ifrit.Runner = (*Router)(nil)

func New(das *das.Das, oas *oas.Oas, sqs *sqs.SQS, opts ...RouterOpt) (*Router, error) {
	if das == nil {
		return nil, fmt.Errorf("aggregator: nil das")
	}

	if oas == nil {
		return nil, fmt.Errorf("aggregator: nil oas")
	}

	r := &Router{}
	for _, opt := range opts {
		opt(r)
	}

	if r.group == "" {
		r.group = "/api/v1/job"
	}

	r.das = das
	r.oas = oas
	r.sqs = sqs

	return r, nil
}

const runner = "Router"

func (r *Router) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	router := gin.Default()

	v1_job := router.Group(r.group)
	{
		v1_job.POST("/create/:type", r.create)
		v1_job.GET("/status", r.status)
		v1_job.GET("/result", r.result)
	}

	wait := make(chan error, 1)
	go func() {
		wait <- router.Run()
	}()

	log.Debug().Fields(f{"runner": runner}).Msg("ready")
	close(ready)
	for {
		select {
		case signal := <-signals:
			log.Debug().Fields(f{"runner": runner, "signal": signal.String()}).Msg("received signal")
			return nil
		case err := <-wait:
			log.Err(err).Fields(f{"runner": runner}).Msg("received error from router")
			return fmt.Errorf("router: received error: %w", err)
		}
	}
}
