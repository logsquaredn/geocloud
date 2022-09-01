package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	// postgres must be imported to inject the postgres driver
	// into the database/sql package.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

//go:embed sql/migrations/*.up.sql
var migrations embed.FS

type Datastore struct {
	db   *sql.DB
	stmt *struct {
		createJob               *sql.Stmt
		createCustomer          *sql.Stmt
		updateJob               *sql.Stmt
		getJobByID              *sql.Stmt
		getJobsBefore           *sql.Stmt
		deleteJob               *sql.Stmt
		getTaskByJobID          *sql.Stmt
		getTaskByType           *sql.Stmt
		getTasksByTypes         *sql.Stmt
		getCustomerByID         *sql.Stmt
		getStorage              *sql.Stmt
		createStorage           *sql.Stmt
		deleteStorage           *sql.Stmt
		updateStorage           *sql.Stmt
		getStorageByCustomerID  *sql.Stmt
		getStorageBefore        *sql.Stmt
		getJobsByCustomerID     *sql.Stmt
		getOutputStorageByJobID *sql.Stmt
		getInputStorageByJobID  *sql.Stmt
	}
}

func New(ctx context.Context, addr string) (*Datastore, error) {
	var (
		d = &Datastore{
			stmt: &struct {
				createJob               *sql.Stmt
				createCustomer          *sql.Stmt
				updateJob               *sql.Stmt
				getJobByID              *sql.Stmt
				getJobsBefore           *sql.Stmt
				deleteJob               *sql.Stmt
				getTaskByJobID          *sql.Stmt
				getTaskByType           *sql.Stmt
				getTasksByTypes         *sql.Stmt
				getCustomerByID         *sql.Stmt
				getStorage              *sql.Stmt
				createStorage           *sql.Stmt
				deleteStorage           *sql.Stmt
				updateStorage           *sql.Stmt
				getStorageByCustomerID  *sql.Stmt
				getStorageBefore        *sql.Stmt
				getJobsByCustomerID     *sql.Stmt
				getOutputStorageByJobID *sql.Stmt
				getInputStorageByJobID  *sql.Stmt
			}{},
		}
		err error
	)
	if d.db, err = sql.Open("postgres", addr); err != nil {
		return nil, err
	}

	if err = d.db.Ping(); err != nil {
		return nil, err
	}

	if d.stmt.createJob, err = d.db.Prepare(createJobSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if d.stmt.createCustomer, err = d.db.Prepare(createCustomerSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if d.stmt.updateJob, err = d.db.Prepare(updateJobSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if d.stmt.getJobByID, err = d.db.Prepare(getJobByIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if d.stmt.getJobsBefore, err = d.db.Prepare(getJobsBeforeSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if d.stmt.deleteJob, err = d.db.Prepare(deleteJobSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if d.stmt.getTaskByJobID, err = d.db.Prepare(getTaskByJobIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if d.stmt.getTaskByType, err = d.db.Prepare(getTaskByTypeSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if d.stmt.getTasksByTypes, err = d.db.Prepare(getTasksByTypesSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if d.stmt.getCustomerByID, err = d.db.Prepare(getCustomerByIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if d.stmt.createStorage, err = d.db.Prepare(createStorageSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if d.stmt.deleteStorage, err = d.db.Prepare(deleteStorageSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if d.stmt.getStorage, err = d.db.Prepare(getStorageByIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if d.stmt.updateStorage, err = d.db.Prepare(updateStorageSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if d.stmt.getJobsByCustomerID, err = d.db.Prepare(getJobsByCustomerIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if d.stmt.getStorageByCustomerID, err = d.db.Prepare(getStorgageByCustomerIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if d.stmt.getStorageBefore, err = d.db.Prepare(getStorageBeforeSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if d.stmt.getOutputStorageByJobID, err = d.db.Prepare(getOutputStorageByJobIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if d.stmt.getInputStorageByJobID, err = d.db.Prepare(getInputStorageByJobIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	return d, nil
}

func (d *Datastore) Close() error {
	return d.db.Close()
}
