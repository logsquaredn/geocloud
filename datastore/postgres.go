package datastore

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	// postgres must be imported to inject the postgres driver
	// into the database/sql package
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/logsquaredn/geocloud"
)

//go:embed psql/migrations/core/*.up.sql
var coreMigrations embed.FS

//go:embed psql/migrations/external/*.up.sql
var externalMigrations embed.FS

type postgresDatastore struct {
	db   *sql.DB
	mgrt *struct {
		core     *migrate.Migrate
		external *migrate.Migrate
	}
	stmt *struct {
		createJob                *sql.Stmt
		createJobCustomerMapping *sql.Stmt
		createCustomer           *sql.Stmt
		updateJob                *sql.Stmt
		getJobByID               *sql.Stmt
		getJobsBefore            *sql.Stmt
		deleteJob                *sql.Stmt
		getTaskByJobID           *sql.Stmt
		getTaskByType            *sql.Stmt
		getTasksByTypes          *sql.Stmt
		getCustomerByCustomerID  *sql.Stmt
	}
}

var _ geocloud.Datastore = (*postgresDatastore)(nil)

func NewPostgres(opts *PostgresDatastoreOpts) (*postgresDatastore, error) {
	var (
		p = &postgresDatastore{
			mgrt: &struct {
				core     *migrate.Migrate
				external *migrate.Migrate
			}{},
			stmt: &struct {
				createJob                *sql.Stmt
				createJobCustomerMapping *sql.Stmt
				createCustomer           *sql.Stmt
				updateJob                *sql.Stmt
				getJobByID               *sql.Stmt
				getJobsBefore            *sql.Stmt
				deleteJob                *sql.Stmt
				getTaskByJobID           *sql.Stmt
				getTaskByType            *sql.Stmt
				getTasksByTypes          *sql.Stmt
				getCustomerByCustomerID  *sql.Stmt
			}{},
		}
		err error
		i   int64 = 1
	)
	for p.db, err = sql.Open("postgres", opts.connectionString()); err != nil; i++ {
		if i >= opts.Retries && opts.Retries > 0 {
			return nil, fmt.Errorf("failed to connect to db after %d attempts: %w", i, err)
		}
		time.Sleep(opts.RetryDelay)
	}

	i = 1
	for err = p.db.Ping(); err != nil; i++ {
		if i >= opts.Retries && opts.Retries > 0 {
			return nil, fmt.Errorf("failed to ping db after %d attempts: %w", i, err)
		}
		time.Sleep(opts.RetryDelay)
	}

	coreSrc, err := iofs.New(coreMigrations, "psql/migrations/core")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations: %w", err)
	}

	for p.mgrt.core, err = migrate.NewWithSourceInstance(
		"core", coreSrc,
		fmt.Sprintf("%s&x-migrations-table=core_migrations", opts.connectionString()),
	); err != nil; i++ {
		if i >= opts.Retries && opts.Retries > 0 {
			return nil, fmt.Errorf("failed to create migrations after %d attempts: %w", i, err)
		}
		time.Sleep(opts.RetryDelay)
	}

	externalSrc, err := iofs.New(externalMigrations, "psql/migrations/external")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations: %w", err)
	}

	for p.mgrt.external, err = migrate.NewWithSourceInstance(
		"external", externalSrc,
		fmt.Sprintf("%s&x-migrations-table=external_migrations", opts.connectionString()),
	); err != nil; i++ {
		if i >= opts.Retries && opts.Retries > 0 {
			return nil, fmt.Errorf("failed to create migrations after %d attempts: %w", i, err)
		}
		time.Sleep(opts.RetryDelay)
	}

	return p, nil
}

func (p *postgresDatastore) Prepare() error {
	var err error

	if p.stmt.createJob, err = p.db.Prepare(createJobSQL); err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.createJobCustomerMapping, err = p.db.Prepare(createJobCustomerMappingSQL); err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.createCustomer, err = p.db.Prepare(createCustomerSQL); err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.updateJob, err = p.db.Prepare(updateJobSQL); err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.getJobByID, err = p.db.Prepare(getJobByIDSQL); err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.getJobsBefore, err = p.db.Prepare(getJobsBeforeSQL); err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.deleteJob, err = p.db.Prepare(deleteJobSQL); err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.getTaskByJobID, err = p.db.Prepare(getTaskByJobIDSQL); err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.getTaskByType, err = p.db.Prepare(getTaskByTypeSQL); err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.getTasksByTypes, err = p.db.Prepare(getTasksByTypesSQL); err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.getCustomerByCustomerID, err = p.db.Prepare(getCustomerByCustomerIDSQL); err != nil {
		return fmt.Errorf("failed to prepare statement; %w", err)
	}

	return nil
}

//go:embed psql/execs/create_job.sql
var createJobSQL string

//go:embed psql/execs/create_job_customer_mapping.sql
var createJobCustomerMappingSQL string

func (p *postgresDatastore) CreateJob(j *geocloud.Job) (*geocloud.Job, error) {
	var (
		id        = uuid.New().String()
		jobErr    sql.NullString
		jobStatus string
		endTime   sql.NullTime
		taskType  string
	)

	err := p.stmt.createJob.QueryRow(
		id, j.TaskType.String(),
		pq.Array(j.Args),
	).Scan(
		&j.ID, &taskType,
		&jobStatus, &jobErr,
		&j.StartTime, &endTime,
		pq.Array(&j.Args),
	)
	if err != nil {
		return j, err
	}

	j.Err = fmt.Errorf(jobErr.String)
	j.EndTime = endTime.Time

	j.TaskType, err = geocloud.TaskTypeFrom(taskType)
	if err != nil {
		return j, err
	}

	j.Status, err = geocloud.JobStatusFrom(jobStatus)
	if err != nil {
		return j, err
	}

	if j.CustomerID != "" {
		err = p.stmt.createJobCustomerMapping.QueryRow(
			id, j.CustomerID,
		).Scan(&j.CustomerID)
		if err != nil {
			return j, err
		}
	}

	return j, nil
}

//go:embed psql/execs/update_job.sql
var updateJobSQL string

func (p *postgresDatastore) UpdateJob(j *geocloud.Job) (*geocloud.Job, error) {
	var (
		jobErr      sql.NullString
		jobStatus   string
		endTime     sql.NullTime
		taskType    string
		jobErrError = ""
	)

	// avoid nil pointer dereference on j.Err.Error()
	if j.Err != nil {
		jobErrError = j.Err.Error()
	}

	err := p.stmt.updateJob.QueryRow(
		j.GetID(), j.TaskType.String(),
		j.Status.String(), jobErrError,
		j.StartTime, j.EndTime,
		pq.Array(j.Args),
	).Scan(
		&j.ID, &taskType,
		&jobStatus, &jobErr,
		&j.StartTime, &endTime,
		pq.Array(&j.Args),
	)
	if err != nil {
		return j, err
	}

	j.Err = fmt.Errorf(jobErr.String)
	j.EndTime = endTime.Time

	j.TaskType, err = geocloud.TaskTypeFrom(taskType)
	if err != nil {
		return j, err
	}

	j.Status, err = geocloud.JobStatusFrom(jobStatus)
	if err != nil {
		return j, err
	}

	return j, nil
}

//go:embed psql/queries/get_job_by_id.sql
var getJobByIDSQL string

func (p *postgresDatastore) GetJob(m geocloud.Message) (*geocloud.Job, error) {
	var (
		j         = &geocloud.Job{}
		jobErr    sql.NullString
		jobStatus string
		endTime   sql.NullTime
		taskType  string
	)

	err := p.stmt.getJobByID.QueryRow(m.GetID()).Scan(
		&j.ID, &taskType,
		&jobStatus, &jobErr,
		&j.StartTime, &endTime,
		pq.Array(&j.Args),
	)
	if err != nil {
		return j, err
	}

	j.Err = fmt.Errorf(jobErr.String)
	j.EndTime = endTime.Time

	j.TaskType, err = geocloud.TaskTypeFrom(taskType)
	if err != nil {
		return j, err
	}

	j.Status, err = geocloud.JobStatusFrom(jobStatus)
	if err != nil {
		return j, err
	}

	return j, nil
}

//go:embed psql/queries/get_jobs_before.sql
var getJobsBeforeSQL string

func (p *postgresDatastore) GetJobs(before time.Duration) ([]*geocloud.Job, error) {
	beforeTimestamp := time.Now().Add(-before)
	rows, err := p.stmt.getJobsBefore.Query(beforeTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*geocloud.Job

	for rows.Next() {
		var (
			j         = &geocloud.Job{}
			jobErr    sql.NullString
			jobStatus string
			endTime   sql.NullTime
			taskType  string
		)

		err = rows.Scan(
			&j.ID, &taskType,
			&jobStatus, &jobErr,
			&j.StartTime, &endTime,
			pq.Array(&j.Args),
			&j.CustomerID,
		)
		if err != nil {
			return nil, err
		}

		j.Err = fmt.Errorf(jobErr.String)
		j.EndTime = endTime.Time

		j.TaskType, err = geocloud.TaskTypeFrom(taskType)
		if err != nil {
			return nil, err
		}

		j.Status, err = geocloud.JobStatusFrom(jobStatus)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, j)
	}

	return jobs, nil
}

//go:embed psql/execs/delete_job.sql
var deleteJobSQL string

func (p *postgresDatastore) DeleteJob(j *geocloud.Job) error {
	_, err := p.stmt.deleteJob.Exec(j.ID)
	if err != nil {
		return err
	}

	return nil
}

//go:embed psql/queries/get_task_by_job_id.sql
var getTaskByJobIDSQL string

func (p *postgresDatastore) GetTaskByJobID(m geocloud.Message) (*geocloud.Task, error) {
	var (
		t        = &geocloud.Task{}
		queueID  sql.NullString
		taskType string
	)

	err := p.stmt.getTaskByJobID.QueryRow(m.GetID()).Scan(&taskType, pq.Array(&t.Params), &queueID)
	if err != nil {
		return t, err
	}

	t.QueueID = queueID.String
	t.Type, err = geocloud.TaskTypeFrom(taskType)
	return t, err
}

//go:embed psql/queries/get_task_by_type.sql
var getTaskByTypeSQL string

func (p *postgresDatastore) GetTask(tt geocloud.TaskType) (*geocloud.Task, error) {
	var (
		t        = &geocloud.Task{}
		queueID  sql.NullString
		taskType string
	)
	err := p.stmt.getTaskByType.QueryRow(tt.String()).Scan(&taskType, pq.Array(&t.Params), &queueID)
	if err != nil {
		return t, err
	}

	t.QueueID = queueID.String
	t.Type, err = geocloud.TaskTypeFrom(taskType)
	return t, err
}

//go:embed psql/queries/get_tasks_by_types.sql
var getTasksByTypesSQL string

func (p *postgresDatastore) GetTasks(tts ...geocloud.TaskType) (ts []*geocloud.Task, err error) {
	ttss := make([]string, len(tts))
	for i, tt := range tts {
		ttss[i] = tt.String()
	}

	rows, err := p.stmt.getTasksByTypes.Query(pq.Array(ttss))
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			task     = &geocloud.Task{}
			queueID  sql.NullString
			taskType string
		)

		err = rows.Scan(&taskType, pq.Array(&task.Params), &queueID)
		if err != nil {
			return
		}

		task.QueueID = queueID.String
		task.Type, err = geocloud.TaskTypeFrom(taskType)
		if err != nil {
			return
		}

		ts = append(ts, task)
	}
	err = rows.Err()
	return
}

//go:embed psql/queries/get_customer_by_customer_id.sql
var getCustomerByCustomerIDSQL string

func (p *postgresDatastore) GetCustomer(customerID string) (*geocloud.Customer, error) {
	c := &geocloud.Customer{}
	err := p.stmt.getCustomerByCustomerID.QueryRow(customerID).Scan(&c.ID, &c.Name)
	if err != nil {
		return c, err
	}

	return c, nil
}

//go:embed psql/execs/create_customer.sql
var createCustomerSQL string

func (p *postgresDatastore) CreateCustomer(customerID string, customer_name string) error {
	_, err := p.stmt.createCustomer.Exec(customerID, customer_name)
	if err != nil {
		return err
	}

	return nil
}

func (p *postgresDatastore) Close() error {
	defer p.stmt.createJob.Close()
	defer p.stmt.createJobCustomerMapping.Close()
	defer p.stmt.updateJob.Close()
	defer p.stmt.getJobByID.Close()
	defer p.stmt.getJobsBefore.Close()
	defer p.stmt.getTaskByJobID.Close()
	defer p.stmt.getTaskByType.Close()
	defer p.stmt.getTasksByTypes.Close()
	defer p.stmt.createJobCustomerMapping.Close()
	return p.db.Close()
}

func (p *postgresDatastore) MigrateCore() error {
	if err := p.mgrt.core.Up(); !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func (p *postgresDatastore) MigrateExternal() error {
	if err := p.mgrt.external.Up(); !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
