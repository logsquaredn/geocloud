package das

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Das struct {
	db      *sql.DB
	retries int
	delay   time.Duration

	stmts struct {
		getStatusById        *sql.Stmt
		getTypeById          *sql.Stmt
		insertNewJob         *sql.Stmt
		getParamsByType      *sql.Stmt
		getQueueNameByType   *sql.Stmt
		getQueueNamesByTypes *sql.Stmt
	}
}

const driver = "postgres"

//go:embed get_job_status_by_job_id.sql
var getStatusByIdSql string

//go:embed get_task_type_by_job_id.sql
var getTypeByIdSql string

//go:embed insert_new_job.sql
var insertNewJobSql string

//go:embed get_task_params_by_task_type.sql
var getParamsByTypeSql string

//go:embed get_task_queue_name_by_task_type.sql
var getQueueNameByTypeSql string

//go:embed get_task_queue_names_by_task_types.sql
var getQueueNamesByTypesSql string

func New(conn string, opts ...DasOpt) (*Das, error) {
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

	d.stmts.getStatusById, err = d.db.Prepare(getStatusByIdSql)
	if err != nil {
		return nil, fmt.Errorf("das: failed to prepare statement: %w", err)
	}

	d.stmts.getTypeById, err = d.db.Prepare(getTypeByIdSql)
	if err != nil {
		return nil, fmt.Errorf("das: failed to prepare statement: %w", err)
	}

	d.stmts.insertNewJob, err = d.db.Prepare(insertNewJobSql)
	if err != nil {
		return nil, fmt.Errorf("das: failed to prepare statement: %w", err)
	}

	d.stmts.getParamsByType, err = d.db.Prepare(getParamsByTypeSql)
	if err != nil {
		return nil, fmt.Errorf("das: failed to prepare statement: %w", err)
	}

	d.stmts.getQueueNameByType, err = d.db.Prepare(getQueueNameByTypeSql)
	if err != nil {
		return nil, err
	}

	d.stmts.getQueueNamesByTypes, err = d.db.Prepare(getQueueNamesByTypesSql)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Das) GetJobStatusByJobId(id string) (jobStatus string, err error) {
	err = d.stmts.getStatusById.QueryRow(id).Scan(&jobStatus)
	return
}

func (d *Das) GetJobTypeByJobId(id string) (jobType string, err error) {
	err = d.stmts.getTypeById.QueryRow(id).Scan(&jobType)
	return
}

func (d *Das) InsertNewJob(id string, jobType string) (err error) {
	_, err = d.stmts.insertNewJob.Exec(id, jobType)
	return
}

func (d *Das) GetTaskParamsByTaskType(taskType string) (params []string, err error) {
	err = d.stmts.getParamsByType.QueryRow(taskType).Scan(pq.Array(&params))
	return
}

func (d *Das) GetQueueNameByTaskType(taskType string) (queueName string, err error) {
	err = d.stmts.getQueueNameByType.QueryRow(taskType).Scan(&queueName)
	return
}

func (d *Das) GetQueueNamesByTaskTypes(tasks... string) (queues []string, err error) {
	rows, err := d.stmts.getQueueNamesByTypes.Query(pq.Array(tasks))
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var queue string
		err = rows.Scan(&queue)
		if err != nil {
			return
		}
		if queue != "" {
			queues = append(queues, queue)
		}
	}
	err = rows.Err()
	return
}

func (d *Das) Close() error {
	d.stmts.getStatusById.Close()
	d.stmts.getTypeById.Close()
	d.stmts.insertNewJob.Close()
	d.stmts.getParamsByType.Close()
	d.stmts.getQueueNameByType.Close()
	d.stmts.getQueueNamesByTypes.Close()
	return d.db.Close()
}
