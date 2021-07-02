package das

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "github.com/lib/pq"
)

type Das struct {
	db   *sql.DB
	conn string

	stmts struct {
		getStatusById *sql.Stmt
		getTypeById *sql.Stmt
	}
}

const driver = "postgres"

//go:embed get_job_status_by_job_id.sql
var getStatusByIdSql string

//go:embed get_job_type_by_job_id.sql
var getTypeByIdSql string

func New(opts ...DasOpt) (*Das, error) {
	d := &Das{}
	for _, opt := range opts {
		opt(d)
	}

	if d.db == nil {
		if d.conn == "" {
			return nil, fmt.Errorf("nil db empty connection string")
		}

		var err error
		d.db, err = sql.Open(driver, d.conn)
		if err != nil {
			return nil, err
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

	return d, nil
}

func (d *Das) GetJobStatusByJobId(id string) (jobStatus string, err error) {
	rows, err := d.stmts.getStatusById.Query(id)
	if err != nil {
		return "", err
	}

	err = rows.Scan(&jobStatus)
	if err != nil {
		return "", err
	}

	return jobStatus, nil
}

func (d *Das) GetJobTypeByJobId(id string) (jobType string, err error) {
	rows, err := d.stmts.getTypeById.Query(id)
	if err != nil {
		return "", err
	}

	err = rows.Scan(&jobType)
	if err != nil {
		return "", err
	}

	return jobType, nil
}

func (d *Das) Close() error {
	d.stmts.getStatusById.Close()
	d.stmts.getTypeById.Close()
	return d.db.Close()
}
