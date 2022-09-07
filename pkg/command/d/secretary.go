package command

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/pkg/store/blob/bucket"
	"github.com/logsquaredn/rototiller/pkg/store/data/postgres"
	"github.com/logsquaredn/rototiller/pkg/volume"
	"github.com/spf13/cobra"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
)

func NewSecretary() *cobra.Command {
	var (
		defaultDuration                             = time.Hour * 24
		workJobsBefore, workStorageBefore           time.Duration
		postgresAddr, bucketAddr, archiveBucketAddr string
		secretaryCmd                                = &cobra.Command{
			Use:     "secretary",
			Aliases: []string{"s"},
			RunE: func(cmd *cobra.Command, args []string) error {
				var (
					ctx     = cmd.Context()
					logr    = rototiller.LoggerFrom(ctx)
					archive strings.Builder
				)

				datastore, err := postgres.New(ctx, postgresAddr)
				if err != nil {
					return err
				}

				blobstore, err := bucket.New(ctx, bucketAddr)
				if err != nil {
					return err
				}

				archiveBlobstore, err := bucket.New(ctx, archiveBucketAddr)
				if err != nil {
					return err
				}

				i := customer.List(&stripe.CustomerListParams{})
				for i.Next() {
					c := i.Customer()
					if _, err := datastore.CreateCustomer(c.ID); err != nil {
						logr.Error(err, "creating customer", "id", c.ID, "name", c.Name)
						return err
					}
				}

				logr.Info("getting jobs")
				jobs, err := datastore.GetJobsBefore(workJobsBefore)
				if err != nil {
					logr.Error(err, "getting jobs")
					return err
				}

				logr.Info("processing jobs")
				for _, j := range jobs {
					c, err := customer.Get(j.GetCustomerId(), nil)
					if err != nil {
						logr.Error(err, "getting customer", "id", j.GetCustomerId())
						return err
					}

					logr.Info("parsing charge_rate for customer", "id", j.GetCustomerId())
					chargeRate, err := strconv.ParseInt(c.Metadata["charge_rate"], 10, 64)
					if err != nil {
						return err
					}

					logr.Info("updating balance for customer", "id", j.GetCustomerId())
					updateBalance := c.Balance + chargeRate
					if _, err = customer.Update(j.GetCustomerId(), &stripe.CustomerParams{
						Balance: &updateBalance,
					}); err != nil {
						logr.Error(err, "updating balance for customer", "id", j.GetCustomerId())
						return err
					}

					logr.Info("deleting data for customer", "id", j.GetCustomerId())
					if err = datastore.DeleteJob(j.GetId()); err != nil {
						logr.Error(err, "deleting data for customer", "id", j.GetCustomerId())
						return err
					}

					archive.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n", j.GetId(), j.GetInputId(), j.GetOutputId(), j.GetTaskType(), j.GetStatus(), j.GetError(), j.GetStartTime().String(), j.GetEndTime().String(), strings.Join(j.GetArgs(), "|"), j.GetCustomerId(), c.Name))
				}

				if len(archive.String()) > 0 {
					logr.Info("putting archive to storage")
					if err = archiveBlobstore.PutObject(
						ctx,
						fmt.Sprint(time.Now().Unix()),
						volume.New(
							volume.NewFile("archive.csv", strings.NewReader(archive.String()), archive.Len()),
						),
					); err != nil {
						logr.Error(err, "putting archive to s3")
						return err
					}
				}

				logr.Info("getting storages")
				storages, err := datastore.GetStorageBefore(workJobsBefore)
				if err != nil {
					logr.Error(err, "getting storages")
					return err
				}

				logr.Info("processing storages")
				for _, s := range storages {
					logr.Info("deleting storage", "id", s.GetId())
					if err = blobstore.DeleteObject(ctx, s.GetId()); err != nil {
						logr.Error(err, "deleting storage", "id", s.GetId())
						return err
					}

					logr.Info("deleting storage data: %s", s.GetId())
					err = datastore.DeleteStorage(s.GetId())
					if err != nil {
						logr.Error(err, "deleting storage data: %s", s.GetId())
					}
				}

				logr.Info("done")

				return nil
			},
		}
	)

	secretaryCmd.Flags().StringVar(&archiveBucketAddr, "archive-bucket-addr", "", "archive bucket address")
	secretaryCmd.Flags().StringVar(&bucketAddr, "bucket-addr", "", "bucket address")
	secretaryCmd.Flags().StringVar(&postgresAddr, "postgres-addr", "", "Postgres address")
	secretaryCmd.Flags().StringVar(&stripe.Key, "stripe-api-key", "", "Stripe API key")
	secretaryCmd.Flags().DurationVar(&workJobsBefore, "work-jobs-before", defaultDuration, "work jobs before")
	secretaryCmd.Flags().DurationVar(&workStorageBefore, "work-storage-before", defaultDuration, "work storage before")

	return secretaryCmd
}
