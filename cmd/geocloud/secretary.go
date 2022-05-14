package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/logsquaredn/geocloud"
	"github.com/logsquaredn/geocloud/datastore"
	"github.com/logsquaredn/geocloud/objectstore"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
)

var secretaryCmd = &cobra.Command{
	Use:     "secretary",
	Aliases: []string{"s"},
	RunE:    runSecretary,
}

var (
	workJobsBefore time.Duration
	stripeAPIKey   string
)

func init() {
	secretaryCmd.Flags().DurationVar(&workJobsBefore, "work-jobs-before", h24, "Work jobs before")
	secretaryCmd.Flags().StringVar(&stripeAPIKey, "stripe-api-key", os.Getenv("GEOCLOUD_STRIPE_API_KEY"), "Work jobs before")
}

func runSecretary(cmd *cobra.Command, args []string) error {
	ds, err := datastore.NewPostgres(
		getPostgresOpts(),
	)
	if err != nil {
		log.Err(err).Msg("creating datastore")
		return err
	}

	if err = ds.Prepare(); err != nil {
		log.Err(err).Msg("preparing datastore")
		return err
	}

	os, err := objectstore.NewS3(
		getS3Opts(),
	)
	if err != nil {
		log.Err(err).Msg("creating object store")
		return err
	}

	osa, err := objectstore.NewS3(
		getS3ArchiveOpts(),
	)
	if err != nil {
		log.Err(err).Msg("creating archive object store")
		return err
	}

	stripe.Key = stripeAPIKey

	log.Info().Msg("importing customers")
	i := customer.List(&stripe.CustomerListParams{})
	for i.Next() {
		c := i.Customer()
		if err := ds.CreateCustomer(&geocloud.Customer{
			ID:   c.ID,
			Name: c.Name,
		}); err != nil {
			log.Err(err).Msgf("creating customer '%s' '%s'", c.ID, c.Name)
			return err
		}
	}

	log.Info().Msg("getting jobs")
	jobs, err := ds.GetJobs(workJobsBefore)
	if err != nil {
		log.Err(err).Msg("getting jobs")
		return err
	}

	var archive strings.Builder
	log.Info().Msg("processing jobs")
	for _, j := range jobs {
		c, err := customer.Get(j.CustomerID, nil)
		if err != nil {
			log.Err(err).Msgf("getting customer '%s'", j.CustomerID)
			return err
		}

		log.Debug().Msgf("parsing charge_rate for customer '%s'", j.CustomerID)
		chargeRate, err := strconv.ParseInt(c.Metadata["charge_rate"], 10, 64)
		if err != nil {
			return err
		}

		log.Debug().Msgf("updating balance for customer '%s'", j.CustomerID)
		updateBalance := c.Balance + chargeRate
		_, err = customer.Update(j.CustomerID, &stripe.CustomerParams{
			Balance: &updateBalance,
		})
		if err != nil {
			log.Err(err).Msgf("updating balance for customer '%s'", j.CustomerID)
			return err
		}

		log.Debug().Msgf("deleting objects for customer '%s'", j.CustomerID)
		err = os.DeleteObject(j.ID)
		if err != nil {
			log.Err(err).Msgf("deleting objects for customer '%s'", j.CustomerID)
			return err
		}

		log.Debug().Msgf("deleting data for customer '%s'", j.CustomerID)
		err = ds.DeleteJob(j)
		if err != nil {
			log.Err(err).Msgf("deleting data for customer '%s'", j.CustomerID)
			return err
		}

		archive.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s\n", j.ID, j.TaskType.String(), j.Status.String(), j.Err.Error(), j.StartTime.String(), j.EndTime.String(), strings.Join(j.Args, "|"), j.CustomerID, c.Name))
	}

	if len(archive.String()) > 0 {
		err = osa.PutObject(
			geocloud.NewMessage(fmt.Sprint(time.Now().Unix())),
			geocloud.NewBytesVolume("archive.csv", []byte(archive.String())),
		)
		if err != nil {
			log.Err(err).Msg("putting archive to s3")
			return err
		}
	}

	log.Info().Msg("done")
	return nil
}
