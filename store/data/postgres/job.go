package postgres

import (
	"database/sql"
	_ "embed"
	"time"

	"github.com/google/uuid"
	"github.com/logsquaredn/rototiller/pb"
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

	//go:embed sql/queries/get_jobs_by_namespace.sql
	getJobsByNamespaceSQL string
)

func (d *Datastore) CreateJob(j *pb.Job) (*pb.Job, error) {
	var (
		id                 = uuid.New().String()
		jobErr             sql.NullString
		startTime, endTime sql.NullTime
		outputID           sql.NullString
	)

	if err := d.stmt.createJob.QueryRow(
		id, j.Namespace,
		j.InputId,
	).Scan(
		&j.Id, &j.Namespace,
		&j.InputId, &outputID,
		&j.Status, &jobErr,
		&startTime, &endTime,
	); err != nil {
		return j, err
	}

	j.Error = jobErr.String
	j.StartTime = timestamppb.New(startTime.Time)
	j.EndTime = timestamppb.New(endTime.Time)
	j.OutputId = outputID.String

	var err error
	j.Steps, err = d.createSteps(j.Id, j.Steps)
	if err != nil {
		return j, err
	}

	return j, nil
}

func (d *Datastore) UpdateJob(j *pb.Job) (*pb.Job, error) {
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
			&j.Id, &j.Namespace,
			&j.InputId, &outputID,
			&j.Status, &jobErr,
			&startTime, &endTime,
		); err != nil {
			return j, err
		}
	} else {
		if err := d.stmt.updateJob.QueryRow(
			j.Id, nil,
			j.Status, j.Error,
			j.StartTime.AsTime(), j.EndTime.AsTime(),
		).Scan(
			&j.Id, &j.Namespace,
			&j.InputId, &outputID,
			&j.Status, &jobErr,
			&startTime, &endTime,
		); err != nil {
			return j, err
		}
	}

	j.Error = jobErr.String
	j.StartTime = timestamppb.New(startTime.Time)
	j.EndTime = timestamppb.New(endTime.Time)
	j.OutputId = outputID.String

	// TODO any reason why Steps would need updated?

	return j, nil
}

func (d *Datastore) GetJob(id string) (*pb.Job, error) {
	var (
		j                  = &pb.Job{}
		jobErr, outputID   sql.NullString
		startTime, endTime sql.NullTime
		err                error
	)

	if err = d.stmt.getJobByID.QueryRow(id).Scan(
		&j.Id, &j.Namespace,
		&j.InputId, &outputID,
		&j.Status, &jobErr,
		&startTime, &endTime,
	); err != nil {
		return j, err
	}

	j.Error = jobErr.String
	j.StartTime = timestamppb.New(startTime.Time)
	j.EndTime = timestamppb.New(endTime.Time)
	j.OutputId = outputID.String

	j.Steps, err = d.getSteps(j.Id)
	if err != nil {
		return j, err
	}

	return j, nil
}

func (d *Datastore) GetJobsBefore(duration time.Duration) ([]*pb.Job, error) {
	beforeTimestamp := time.Now().Add(-duration)
	rows, err := d.stmt.getJobsBefore.Query(beforeTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*pb.Job

	for rows.Next() {
		var (
			j                  = &pb.Job{}
			jobErr             sql.NullString
			startTime, endTime sql.NullTime
			outputID           sql.NullString
		)

		err = rows.Scan(
			&j.Id, &j.Namespace,
			&j.InputId, &outputID,
			&j.Status, &jobErr,
			&startTime, &endTime,
		)
		if err != nil {
			return nil, err
		}

		j.Error = jobErr.String
		j.StartTime = timestamppb.New(startTime.Time)
		j.EndTime = timestamppb.New(endTime.Time)
		j.OutputId = outputID.String

		j.Steps, err = d.getSteps(j.Id)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, j)
	}

	return jobs, nil
}

func (d *Datastore) DeleteJob(id string) error {
	_, err := d.stmt.deleteJob.Exec(id)
	return err
}

func (d *Datastore) GetJobs(id string, offset, limit int) ([]*pb.Job, error) {
	rows, err := d.stmt.getJobsByNamespace.Query(id, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := []*pb.Job{}
	for rows.Next() {
		var (
			j                  = &pb.Job{}
			jobErr, outputID   sql.NullString
			startTime, endTime sql.NullTime
		)

		err = rows.Scan(
			&j.Id, &j.Namespace,
			&j.InputId, &outputID,
			&j.Status, &jobErr,
			&startTime, &endTime,
		)
		if err != nil {
			return nil, err
		}

		j.Error = jobErr.String
		j.StartTime = timestamppb.New(startTime.Time)
		j.EndTime = timestamppb.New(endTime.Time)
		j.OutputId = outputID.String

		j.Steps, err = d.getSteps(j.Id)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, j)
	}

	return jobs, nil
}
