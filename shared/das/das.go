package das

import (
	"database/sql"
	_ "embed"

	"github.com/lib/pq"
)

type Das struct {
	db      *sql.DB
	retries int

	stmts struct {
		getStatusById      *sql.Stmt
		getTypeById        *sql.Stmt
		insertNewJob       *sql.Stmt
		getParamsByType    *sql.Stmt
		getQueueNameByType *sql.Stmt
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

func New(conn string, opts ...DasOpt) (*Das, error) {
	d := &Das{}
	for _, opt := range opts {
		opt(d)
	}

	var err error
	d.db, err = sql.Open(driver, conn)
	if err != nil {
		return nil, err
	}

	err = d.db.Ping()
	if err != nil {
		return nil, err
	}

	d.stmts.getStatusById, err = d.db.Prepare(getStatusByIdSql)
	if err != nil {
		return nil, err
	}

	d.stmts.getTypeById, err = d.db.Prepare(getTypeByIdSql)
	if err != nil {
		return nil, err
	}

	d.stmts.insertNewJob, err = d.db.Prepare(insertNewJobSql)
	if err != nil {
		return nil, err
	}

	d.stmts.getParamsByType, err = d.db.Prepare(getParamsByTypeSql)
	if err != nil {
		return nil, err
	}

	d.stmts.getQueueNameByType, err = d.db.Prepare(getQueueNameByTypeSql)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Das) GetJobStatusByJobId(id string) (jobStatus string, err error) {
	err = d.stmts.getStatusById.QueryRow(id).Scan(&jobStatus)
	if err != nil {
		return "", err
	}

	return jobStatus, nil
}

func (d *Das) GetJobTypeByJobId(id string) (jobType string, err error) {
	err = d.stmts.getTypeById.QueryRow(id).Scan(&jobType)
	if err != nil {
		return "", err
	}

	return jobType, nil
}

func (d *Das) InsertNewJob(id string, jobType string) (err error) {
	_, err = d.stmts.insertNewJob.Exec(id, jobType)

	return err
}

func (d *Das) GetTaskParamsByTaskType(taskType string) (params []string, err error) {
	err = d.stmts.getParamsByType.QueryRow(taskType).Scan(pq.Array(&params))
	if err != nil {
		return nil, err
	}

	return params, nil
}

func (d *Das) GetQueueNameByTaskType(taskType string) (queueName string, err error) {
	err = d.stmts.getQueueNameByType.QueryRow(taskType).Scan(&queueName)
	if err != nil {
		return "", err
	}

	return queueName, nil
}

func (d *Das) Close() error {
	d.stmts.getStatusById.Close()
	d.stmts.getTypeById.Close()
	return d.db.Close()
}
