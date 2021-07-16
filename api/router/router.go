package router

import (
	"database/sql"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud/shared/das"
	"github.com/logsquaredn/geocloud/shared/oas"
	"github.com/rs/zerolog/log"
	"github.com/tedsuo/ifrit"
)

type f = map[string]interface{}

type Router struct {
	conn  string
	das   *das.Das
	oas   *oas.Oas
	db    *sql.DB
	group string
}

var _ ifrit.Runner = (*Router)(nil)

func New(opts ...RouterOpt) (*Router, error) {
	r := &Router{}
	for _, opt := range opts {
		opt(r)
	}

	if r.group == "" {
		r.group = "/api/v1/job"
	}

	if r.das == nil {
		opts := []das.DasOpt{
			das.WithConnectionString(r.conn),
			das.WithDB(r.db),
		}

		var err error
		r.das, err = das.New(opts...)
		if err != nil {
			return nil, err
		}
	}

	if r.oas == nil {
		opts := []oas.OasOpt{
			oas.WithBucket(""),
			oas.WithRegion(""),
		}

		var err error
		r.oas, err = oas.New(opts...)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

const runner = "Router"

func (r *Router) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	router := gin.Default()

	v1_job := router.Group(r.group)
	{
		v1_job.POST("/create/:type", r.create)
		v1_job.GET("/status", r.status)
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
			return err
		}
	}
}
