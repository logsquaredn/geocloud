package das

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Das struct {
	db      *sql.DB
	conn    string
	retries int

	stmts struct {
		getStatusById   *sql.Stmt
		getTypeById     *sql.Stmt
		insertNewJob    *sql.Stmt
		getParamsByType *sql.Stmt
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

func New(opts ...DasOpt) (*Das, error) {
	d := &Das{}
	for _, opt := range opts {
		opt(d)
	}

	if d.db == nil {
		if d.conn == "" {
			return nil, fmt.Errorf("das: nil db, empty connection string")
		}

		if d.retries == 0 {
			d.retries = 5
		}

		var (
			err error
			i = 0
		)
		for d.db, err = sql.Open(driver, d.conn); err != nil; i++ {
			if i > d.retries {
				return nil, fmt.Errorf("das: failed to connect to DB after %d attempts: %w", d.retries, err)
			}
			time.Sleep(time.Second*10)
		}
	}

	err := d.db.Ping()
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

func (d *Das) Close() error {
	d.stmts.getStatusById.Close()
	d.stmts.getTypeById.Close()
	return d.db.Close()
}
