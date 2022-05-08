package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/logsquaredn/geocloud/datastore"
	"github.com/logsquaredn/geocloud/objectstore"
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

	i := customer.List(&stripe.CustomerListParams{})
	for i.Next() {
		c := i.Customer()
		err := ds.CreateCustomer(c.ID, c.Name)
		if err != nil {
			return err
		}
	}

	jobs, err := ds.GetJobs(workJobsBefore)
	if err != nil {
		return err
	}

	for _, j := range jobs {
		c, err := customer.Get(j.CustomerID, nil)
		if err != nil {
			return err
		}

		chargeRate, err := strconv.ParseInt(c.Metadata["charge_rate"], 10, 64)
		if err != nil {
			return err
		}
		updateBalance := c.Balance + chargeRate
		_, err = customer.Update(j.CustomerID, &stripe.CustomerParams{
			Balance: &updateBalance,
		})
		if err != nil {
			return err
		}

		err = os.DeleteRecursive(fmt.Sprintf("jobs/%s", j.ID))
		if err != nil {
			return err
		}

		err = ds.DeleteJob(j)
		if err != nil {
			return err
		}
	}

	return nil
}
