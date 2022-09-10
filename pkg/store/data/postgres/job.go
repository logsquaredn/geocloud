package postgres

import (
	"database/sql"
	_ "embed"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/logsquaredn/rototiller/pkg/api"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	//go:embed sql/execs/create_job.sql
	createJobSQL string

	//go:embed sql/execs/delete_job.sql
	deleteJobSQL string

	//go:embed sql/execs/update_job.sql
	updateJobSQL string

	//go:embed sql/queries/get_jobs_before.sql
	getJobsBeforeSQL string

	//go:embed sql/queries/get_job_by_id.sql
	getJobByIDSQL string

	//go:embed sql/queries/get_job_by_customer_id.sql
	getJobsByCustomerIDSQL string
)

func (d *Datastore) CreateJob(j *api.Job) (*api.Job, error) {
	var (
		id                 = uuid.New().String()
		jobErr             sql.NullString
		startTime, endTime sql.NullTime
		outputID           sql.NullString
	)

	if err := d.stmt.createJob.QueryRow(
		id, j.CustomerId,
		j.InputId,
		j.TaskType,
		pq.Array(j.Args),
	).Scan(
		&j.Id, &j.CustomerId,
		&j.InputId, &outputID,
		&j.TaskType,
		&j.Status, &jobErr,
		&startTime, &endTime,
		pq.Array(&j.Args),
	); err != nil {
		return j, err
	}

	j.Error = jobErr.String
	j.StartTime = timestamppb.New(startTime.Time)
	j.EndTime = timestamppb.New(endTime.Time)
	j.OutputId = outputID.String

	return j, nil
}

func (d *Datastore) UpdateJob(j *api.Job) (*api.Job, error) {
	var (
		jobErr             sql.NullString
		startTime, endTime sql.NullTime
		outputID           sql.NullString
	)

	if j.OutputId != "" {
		if err := d.stmt.updateJob.QueryRow(
			j.Id, j.OutputId,
			j.Status, j.Error,
			j.StartTime.AsTime(), j.EndTime.AsTime(),
		).Scan(
			&j.Id, &j.CustomerId,
			&j.InputId, &outputID,
			&j.TaskType,
			&j.Status, &jobErr,
			&startTime, &endTime,
			pq.Array(&j.Args),
		); err != nil {
			return j, err
		}
	} else {
		if err := d.stmt.updateJob.QueryRow(
			j.Id, nil,
			j.Status, j.Error,
			j.StartTime.AsTime(), j.EndTime.AsTime(),
		).Scan(
			&j.Id, &j.CustomerId,
			&j.InputId, &outputID,
			&j.TaskType,
			&j.Status, &jobErr,
			&startTime, &endTime,
			pq.Array(&j.Args),
		); err != nil {
			return j, err
		}
	}

	j.Error = jobErr.String
	j.StartTime = timestamppb.New(startTime.Time)
	j.EndTime = timestamppb.New(endTime.Time)
	j.OutputId = outputID.String

	return j, nil
}

func (d *Datastore) GetJob(id string) (*api.Job, error) {
	var (
		j                  = &api.Job{}
		jobErr, outputID   sql.NullString
		startTime, endTime sql.NullTime
		err                error
	)

	if err = d.stmt.getJobByID.QueryRow(id).Scan(
		&j.Id, &j.CustomerId,
		&j.InputId, &outputID,
		&j.TaskType,
		&j.Status, &jobErr,
		&startTime, &endTime,
		pq.Array(&j.Args),
	); err != nil {
		return j, err
	}

	j.Error = jobErr.String
	j.StartTime = timestamppb.New(startTime.Time)
	j.EndTime = timestamppb.New(endTime.Time)
	j.OutputId = outputID.String

	return j, nil
}

func (d *Datastore) GetJobsBefore(duration time.Duration) ([]*api.Job, error) {
	beforeTimestamp := time.Now().Add(-duration)
	rows, err := d.stmt.getJobsBefore.Query(beforeTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*api.Job

	for rows.Next() {
		var (
			j                  = &api.Job{}
			jobErr             sql.NullString
			startTime, endTime sql.NullTime
			outputID           sql.NullString
		)

		err = rows.Scan(
			&j.Id, &j.CustomerId,
			&j.InputId, &outputID,
			&j.TaskType,
			&j.Status, &jobErr,
			&startTime, &endTime,
			pq.Array(&j.Args),
		)
		if err != nil {
			return nil, err
		}

		j.Error = jobErr.String
		j.StartTime = timestamppb.New(startTime.Time)
		j.EndTime = timestamppb.New(endTime.Time)
		j.OutputId = outputID.String

		jobs = append(jobs, j)
	}

	return jobs, nil
}

func (d *Datastore) DeleteJob(id string) error {
	_, err := d.stmt.deleteJob.Exec(id)
	return err
}

func (d *Datastore) GetCustomerJobs(id string, offset, limit int) ([]*api.Job, error) {
	rows, err := d.stmt.getJobsByCustomerID.Query(id, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := []*api.Job{}
	for rows.Next() {
		var (
			j                  = &api.Job{}
			jobErr, outputID   sql.NullString
			startTime, endTime sql.NullTime
		)

		err = rows.Scan(
			&j.Id, &j.CustomerId,
			&j.InputId, &outputID,
			&j.TaskType,
			&j.Status, &jobErr,
			&startTime, &endTime,
			pq.Array(&j.Args),
		)
		if err != nil {
			return nil, err
		}

		j.Error = jobErr.String
		j.StartTime = timestamppb.New(startTime.Time)
		j.EndTime = timestamppb.New(endTime.Time)
		j.OutputId = outputID.String

		jobs = append(jobs, j)
	}

	return jobs, nil
}
