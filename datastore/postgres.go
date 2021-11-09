package datastore

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	// postgres must be imported to inject the postgres driver
	// into the database/sql module
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/logsquaredn/geocloud"
	"github.com/tedsuo/ifrit"
)

type PostgresDatastore struct {
	Host       string        `long:"host" description:"Postgres host"`
	Port       int64         `long:"port" default:"5432" description:"Postgres port"`
	User       string        `long:"user" default:"geocloud" description:"Postgres username"`
	Password   string        `long:"password" description:"Postgres password"`
	SSLMode    string        `long:"ssl-mode" default:"disable" choice:"disable" description:"Postgres SSL mode"`
	Retries    int64         `long:"retries" default:"5" description:"Number of times to retry connecting to Postgres. 0 is infinity"`
	RetryDelay time.Duration `long:"retry-delay" default:"5s" description:"Time to wait between attempts at connecting to Postgres"`

	db   *sql.DB
	stmt struct {
		createJob       *sql.Stmt
		updateJob       *sql.Stmt
		getJobByID      *sql.Stmt
		getTaskByJobID  *sql.Stmt
		getTaskByType   *sql.Stmt
		getTasksByTypes *sql.Stmt
	}
}

var _ geocloud.Datastore = (*PostgresDatastore)(nil)

func (p *PostgresDatastore) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	var (
		err error
		i   int64 = 1
	)
	for p.db, err = sql.Open("postgres", p.connectionString()); err != nil; i++ {
		if i >= p.Retries && p.Retries > 0 {
			return fmt.Errorf("failed to connect to db after %d attempts: %w", i, err)
		}
		time.Sleep(p.RetryDelay)
	}

	i = 1
	for err = p.db.Ping(); err != nil; i++ {
		if i >= p.Retries && p.Retries > 0 {
			return fmt.Errorf("failed to ping db after %d attempts: %w", i, err)
		}
		time.Sleep(p.RetryDelay)
	}

	if p.stmt.createJob, err = p.db.Prepare(createJobSQL); err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.updateJob, err = p.db.Prepare(updateJobSQL); err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.getJobByID, err = p.db.Prepare(getJobByIDSQL); err != nil {
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

	defer p.close()
	close(ready)
	<-signals
	return nil
}

func (p *PostgresDatastore) Execute(_ []string) error {
	return <-ifrit.Invoke(p).Wait()
}

//go:embed psql/execs/create_job.sql
var createJobSQL string

func (p *PostgresDatastore) CreateJob(j *geocloud.Job) (*geocloud.Job, error) {
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
		&j.Id, &taskType,
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

//go:embed psql/execs/update_job.sql
var updateJobSQL string

func (p *PostgresDatastore) UpdateJob(j *geocloud.Job) (*geocloud.Job, error) {
	var (
		jobErr    sql.NullString
		jobStatus string
		endTime   sql.NullTime
		taskType  string
	)

	// avoid nil pointer dereference on j.Err.Error()
	if j.Err == nil {
		j.Err = fmt.Errorf("")
	}

	err := p.stmt.updateJob.QueryRow(
		j.ID(), j.TaskType.String(),
		j.Status.String(), j.Err.Error(),
		j.StartTime, j.EndTime,
		pq.Array(j.Args),
	).Scan(
		&j.Id, &taskType,
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

func (p *PostgresDatastore) GetJob(m geocloud.Message) (*geocloud.Job, error) {
	var (
		j         = &geocloud.Job{}
		jobErr    sql.NullString
		jobStatus string
		endTime   sql.NullTime
		taskType  string
	)

	err := p.stmt.getJobByID.QueryRow(m.ID()).Scan(
		&j.Id, &taskType,
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

//go:embed psql/queries/get_task_by_job_id.sql
var getTaskByJobIDSQL string

func (p *PostgresDatastore) GetTaskByJobID(m geocloud.Message) (*geocloud.Task, error) {
	var (
		t        = &geocloud.Task{}
		taskType string
	)

	err := p.stmt.getTaskByJobID.QueryRow(m.ID()).Scan(&taskType, pq.Array(&t.Params), &t.QueueID, &t.Ref)
	if err != nil {
		return t, err
	}

	t.Type, err = geocloud.TaskTypeFrom(taskType)
	return t, err
}

//go:embed psql/queries/get_task_by_type.sql
var getTaskByTypeSQL string

func (p *PostgresDatastore) GetTask(tt geocloud.TaskType) (*geocloud.Task, error) {
	var (
		t        = &geocloud.Task{}
		taskType string
	)
	err := p.stmt.getTaskByType.QueryRow(tt.String()).Scan(&taskType, pq.Array(&t.Params), &t.QueueID, &t.Ref)
	if err != nil {
		return t, err
	}

	t.Type, err = geocloud.TaskTypeFrom(taskType)
	return t, err
}

//go:embed psql/queries/get_tasks_by_types.sql
var getTasksByTypesSQL string

func (p *PostgresDatastore) GetTasks(tts ...geocloud.TaskType) (ts []*geocloud.Task, err error) {
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
			taskType string
		)

		err = rows.Scan(&taskType, pq.Array(&task.Params), &task.QueueID, &task.Ref)
		if err != nil {
			return
		}

		task.Type, err = geocloud.TaskTypeFrom(taskType)
		if err != nil {
			return
		}

		ts = append(ts, task)
	}
	err = rows.Err()
	return
}

func (p *PostgresDatastore) connectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d?sslmode=%s", p.User, p.Password, p.Host, p.Port, p.SSLMode)
}

func (p *PostgresDatastore) close() error {
	defer p.stmt.createJob.Close()
	defer p.stmt.updateJob.Close()
	defer p.stmt.getJobByID.Close()
	defer p.stmt.getTaskByJobID.Close()
	defer p.stmt.getTaskByType.Close()
	defer p.stmt.getTasksByTypes.Close()
	return p.db.Close()
}

func (p *PostgresDatastore) Name() string {
	return "postgres"
}

func (p *PostgresDatastore) IsConfigured() bool {
	return p != nil && p.Port > 0 && p.Host != "" && p.User != ""
}

func (p *PostgresDatastore) WithDB(db *sql.DB) *PostgresDatastore {
	p.db = db
	return p
}

//go:embed psql/migrations/*.up.sql
var migrations embed.FS

func (p *PostgresDatastore) Migrate() error {
	src, err := iofs.New(migrations, "psql/migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	var (
		m *migrate.Migrate
		i int64 = 1
	)
	for m, err = migrate.NewWithSourceInstance(
		"migrations", src,
		p.connectionString(),
	); err != nil; i++ {
		if i >= p.Retries && p.Retries > 0 {
			return fmt.Errorf("failed to apply migrations after %d attempts: %w", i, err)
		}
		time.Sleep(p.RetryDelay)
	}

	return m.Up()
}
