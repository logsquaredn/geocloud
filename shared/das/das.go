package das

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/logsquaredn/geocloud"
)

type Das struct {
	db      *sql.DB
	retries int
	delay   time.Duration

	stmts struct {
		insertJob                    *sql.Stmt
		getJobByJobID                *sql.Stmt
		getTaskByJobID               *sql.Stmt
		getTaskByTaskType            *sql.Stmt
		getTaskQueueNamesByTaskTypes *sql.Stmt
	}
}

const driver = "postgres"

//go:embed execs/insert_job.sql
var insertJobSQL string

//go:embed queries/get_job_by_job_id.sql
var getJobByJobIDSQL string

//go:embed queries/get_task_by_job_id.sql
var getTaskByJobIDSQL string

//go:embed queries/get_task_by_task_type.sql
var getTaskByTaskTypeSQL string

//go:embed queries/get_task_queue_names_by_task_types.sql
var getTaskQueueNamesByTaskTypesSQL string

func New(conn string, opts... DasOpt) (*Das, error) {
	d := &Das{}
	for _, opt := range opts {
		opt(d)
	}

	var (
		err error
		i = 1
	)
	for d.db, err = sql.Open(driver, conn); err != nil; i++ {
		if i >= d.retries {
			return nil, fmt.Errorf("das: failed to connect to db after %d attempts: %w", i, err)
		}
		time.Sleep(d.delay)
	}

	i = 1
	for err = d.db.Ping(); err != nil; i++ {
		if i >= d.retries {
			return nil, fmt.Errorf("das: failed to ping db after %d attempts: %w", i, err)
		}
		time.Sleep(d.delay)
	}

	if d.stmts.insertJob, err = d.db.Prepare(insertJobSQL); err != nil {
		return nil, fmt.Errorf("das: failed to prepare statement: %w", err)
	}

	if d.stmts.getJobByJobID, err = d.db.Prepare(getJobByJobIDSQL); err != nil {
		return nil, fmt.Errorf("das: failed to prepare statement: %w", err)
	}

	if d.stmts.getTaskByJobID, err = d.db.Prepare(getTaskByJobIDSQL); err != nil {
		return nil, fmt.Errorf("das: failed to prepare statement: %w", err)
	}

	if d.stmts.getTaskByTaskType, err = d.db.Prepare(getTaskByTaskTypeSQL); err != nil {
		return nil, fmt.Errorf("das: failed to prepare statement: %w", err)
	}

	if d.stmts.getTaskQueueNamesByTaskTypes, err = d.db.Prepare(getTaskQueueNamesByTaskTypesSQL); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Das) InsertJob(taskType string) (j geocloud.Job, err error) {
	jobID := uuid.New().String()
	var jobErr string
	err = d.stmts.insertJob.QueryRow(jobID, taskType).Scan(&j.ID, &j.TaskType, &j.Status, &jobErr)
	j.Error = fmt.Errorf(jobErr)
	return
}

func (d *Das) GetJobByJobID(jobID string) (j geocloud.Job, err error) {
	var jobErr string
	err = d.stmts.getJobByJobID.QueryRow(jobID).Scan(&j.ID, &j.TaskType, &j.Status, &jobErr)
	j.Error = fmt.Errorf(jobErr)
	return
}

func (d *Das) GetTaskByTaskType(taskType string) (t geocloud.Task, err error) {
	err = d.stmts.getTaskByTaskType.QueryRow(taskType).Scan(&t.Type, pq.Array(&t.Params), &t.QueueName, &t.Ref)
	return
}

func (d *Das) GetTaskByJobID(jobID string) (t geocloud.Task, err error) {
	err = d.stmts.getTaskByJobID.QueryRow(jobID).Scan(&t.Type, pq.Array(&t.Params), &t.QueueName, &t.Ref)
	return
}

func (d *Das) GetQueueNamesByTaskTypes(taskTypes... string) (queueNames []string, err error) {
	rows, err := d.stmts.getTaskQueueNamesByTaskTypes.Query(pq.Array(taskTypes))
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var queueName string
		err = rows.Scan(&queueName)
		if err != nil {
			return
		}
		if queueName != "" {
			queueNames = append(queueNames, queueName)
		}
	}
	err = rows.Err()
	return
}

func (d *Das) Close() error {
	d.stmts.insertJob.Close()
	d.stmts.getJobByJobID.Close()
	d.stmts.getTaskByJobID.Close()
	d.stmts.getTaskByTaskType.Close()
	d.stmts.getTaskQueueNamesByTaskTypes.Close()
	return d.db.Close()
}
