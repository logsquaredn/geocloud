package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/logsquaredn/geocloud"
	"github.com/logsquaredn/geocloud/datastore"
	"github.com/logsquaredn/geocloud/objectstore"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
)

var secretaryCmd = &cobra.Command{
	Use:     "secretary",
	Aliases: []string{"s"},
	RunE:    runSecretary,
}

var (
	workJobsBefore    time.Duration
	workStorageBefore time.Duration
)

func init() {
	flags := secretaryCmd.Flags()
	flags.DurationVar(&workJobsBefore, "work-jobs-before", h24, "Work jobs before")
	flags.DurationVar(&workStorageBefore, "work-storage-before", h24, "Work storage before")
	flags.String("stripe-api-key", "", "Stripe API key")
	viper.BindPFlag("stripe-api-key", flags.Lookup("stripe-api-key"))
}

func runSecretary(cmd *cobra.Command, args []string) error {
	ds, err := datastore.NewPostgres(
		getPostgresOpts(),
	)
	if err != nil {
		log.Err(err).Msg("creating datastore")
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

	stripe.Key = viper.GetString("stripe-api-key")

	log.Info().Msg("importing customers")
	i := customer.List(&stripe.CustomerListParams{})
	for i.Next() {
		c := i.Customer()
		if _, err := ds.CreateCustomer(&geocloud.Customer{
			ID: c.ID,
		}); err != nil {
			log.Err(err).Msgf("creating customer '%s' '%s'", c.ID, c.Name)
			return err
		}
	}

	log.Info().Msg("getting jobs")
	jobs, err := ds.GetJobsBefore(workJobsBefore)
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

		log.Debug().Msgf("deleting data for customer '%s'", j.CustomerID)
		err = ds.DeleteJob(j)
		if err != nil {
			log.Err(err).Msgf("deleting data for customer '%s'", j.CustomerID)
			return err
		}

		archive.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n", j.ID, j.InputID, j.OutputID, j.TaskType.String(), j.Status.String(), j.Error, j.StartTime.String(), j.EndTime.String(), strings.Join(j.Args, "|"), j.CustomerID, c.Name))
	}

	if len(archive.String()) > 0 {
		log.Info().Msg("putting archive to storage")
		err = osa.PutObject(
			geocloud.NewMessage(fmt.Sprint(time.Now().Unix())),
			geocloud.NewSingleFileVolume("archive.csv", []byte(archive.String())),
		)
		if err != nil {
			log.Err(err).Msg("putting archive to s3")
			return err
		}
	}

	log.Info().Msg("getting storages")
	storages, err := ds.GetStorageBefore(workJobsBefore)
	if err != nil {
		log.Err(err).Msg("getting storages")
		return err
	}

	log.Info().Msg("processing storages")
	for _, s := range storages {
		log.Debug().Msgf("deleting storage: %s", s.ID)
		err = os.DeleteObjectRecursive(s.ID)
		if err != nil {
			log.Err(err).Msgf("deleting storage: %s", s.ID)
			return err
		}

		log.Debug().Msgf("deleting storage data: %s", s.ID)
		err = ds.DeleteStorage(s)
		if err != nil {
			log.Err(err).Msgf("deleting storage data: %s", s.ID)
		}
	}

	log.Info().Msg("done")
	return nil
}
