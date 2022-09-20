package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"os"
	"strings"

	// postgres must be imported to inject the postgres driver
	// into the database/sql package.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

//go:embed sql/migrations/*.up.sql
var migrations embed.FS

type Datastore struct {
	*sql.DB
	stmt *struct {
		createJob               *sql.Stmt
		updateJob               *sql.Stmt
		getJobByID              *sql.Stmt
		getJobsBefore           *sql.Stmt
		deleteJob               *sql.Stmt
		getTaskByJobID          *sql.Stmt
		getTaskByType           *sql.Stmt
		getTasksByTypes         *sql.Stmt
		getStorage              *sql.Stmt
		createStorage           *sql.Stmt
		deleteStorage           *sql.Stmt
		updateStorage           *sql.Stmt
		getStorageByOwnerID     *sql.Stmt
		getStorageBefore        *sql.Stmt
		getJobsByOwnerID        *sql.Stmt
		getOutputStorageByJobID *sql.Stmt
		getInputStorageByJobID  *sql.Stmt
		createSink              *sql.Stmt
		getSinksByJobID         *sql.Stmt
	}
}

func New(ctx context.Context, addr string) (*Datastore, error) {
	d := &Datastore{
		stmt: &struct {
			createJob               *sql.Stmt
			updateJob               *sql.Stmt
			getJobByID              *sql.Stmt
			getJobsBefore           *sql.Stmt
			deleteJob               *sql.Stmt
			getTaskByJobID          *sql.Stmt
			getTaskByType           *sql.Stmt
			getTasksByTypes         *sql.Stmt
			getStorage              *sql.Stmt
			createStorage           *sql.Stmt
			deleteStorage           *sql.Stmt
			updateStorage           *sql.Stmt
			getStorageByOwnerID     *sql.Stmt
			getStorageBefore        *sql.Stmt
			getJobsByOwnerID        *sql.Stmt
			getOutputStorageByJobID *sql.Stmt
			getInputStorageByJobID  *sql.Stmt
			createSink              *sql.Stmt
			getSinksByJobID         *sql.Stmt
		}{},
	}

	if addr == "" {
		addr = os.Getenv("POSTGRES_ADDR")
	}

	addr = "postgres://" + strings.TrimPrefix(addr, "postgres://")

	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	q := u.Query()

	for queryParam, envVar := range map[string]string{
		"sslmode": "POSTGRES_SSLMODE",
	} {
		if value := os.Getenv(envVar); value != "" {
			q.Add(queryParam, value)
		}
	}

	u.RawQuery = q.Encode()

	if u.User.String() == "" {
		u.User = url.UserPassword(os.Getenv("POSTGRES_USERNAME"), os.Getenv("POSTGRES_PASSWORD"))
	}

	if d.DB, err = sql.Open("postgres", u.String()); err != nil {
		return nil, err
	}

	if err = d.DB.Ping(); err != nil {
		return nil, err
	}

	if d.stmt.createJob, err = d.DB.Prepare(createJobSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if d.stmt.updateJob, err = d.DB.Prepare(updateJobSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if d.stmt.getJobByID, err = d.DB.Prepare(getJobByIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if d.stmt.getJobsBefore, err = d.DB.Prepare(getJobsBeforeSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if d.stmt.deleteJob, err = d.DB.Prepare(deleteJobSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if d.stmt.getTaskByJobID, err = d.DB.Prepare(getTaskByJobIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if d.stmt.getTaskByType, err = d.DB.Prepare(getTaskByTypeSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if d.stmt.getTasksByTypes, err = d.DB.Prepare(getTasksByTypesSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if d.stmt.createStorage, err = d.DB.Prepare(createStorageSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if d.stmt.deleteStorage, err = d.DB.Prepare(deleteStorageSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if d.stmt.getStorage, err = d.DB.Prepare(getStorageByIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if d.stmt.updateStorage, err = d.DB.Prepare(updateStorageSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if d.stmt.getJobsByOwnerID, err = d.DB.Prepare(getJobsByOwnerIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if d.stmt.getStorageByOwnerID, err = d.DB.Prepare(getStorgageByOwnerIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if d.stmt.getStorageBefore, err = d.DB.Prepare(getStorageBeforeSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if d.stmt.getOutputStorageByJobID, err = d.DB.Prepare(getOutputStorageByJobIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if d.stmt.getInputStorageByJobID, err = d.DB.Prepare(getInputStorageByJobIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if d.stmt.createSink, err = d.DB.Prepare(createSinkSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if d.stmt.getSinksByJobID, err = d.DB.Prepare(getSinksByJobIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	return d, nil
}
