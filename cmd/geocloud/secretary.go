package main

import (
	"os"
	"strconv"
	"time"

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
		return err
	}

	if err = ds.Prepare(); err != nil {
		return err
	}

	os, err := objectstore.NewS3(
		getS3Opts(),
	)
	if err != nil {
		return err
	}

	stripe.Key = stripeAPIKey

	log.Info().Msg("importing customers")
	i := customer.List(&stripe.CustomerListParams{})
	for i.Next() {
		c := i.Customer()
		if err := ds.CreateCustomer(c.ID, c.Name); err != nil {
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
		err = os.DeleteRecursive(j.ID)
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
	}

	log.Info().Msg("done")
	return nil
}
