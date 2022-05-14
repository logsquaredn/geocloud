package datastore

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type PostgresMigrations struct {
	mgrt *migrate.Migrate
}

func NewPostgresMigrations(opts *PostgresOpts) (*PostgresMigrations, error) {
	var (
		p   = &PostgresMigrations{}
		err error
		i   int64 = 1
	)
	src, err := iofs.New(migrations, "psql/migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations: %w", err)
	}

	for p.mgrt, err = migrate.NewWithSourceInstance(
		"migrations", src,
		opts.connectionString(),
	); err != nil; i++ {
		if i >= opts.Retries && opts.Retries > 0 {
			return nil, fmt.Errorf("failed to create migrations after %d attempts: %w", i, err)
		}
		time.Sleep(opts.RetryDelay)
	}

	return p, nil
}

func (p *PostgresMigrations) Migrate() error {
	if err := p.mgrt.Up(); !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
