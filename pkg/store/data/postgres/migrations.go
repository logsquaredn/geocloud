package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type Migrations struct {
	*migrate.Migrate
}

func NewMigrations(ctx context.Context, addr string) (*Migrations, error) {
	var (
		p   = &Migrations{}
		err error
	)
	src, err := iofs.New(migrations, "sql/migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations: %w", err)
	}

	if p.Migrate, err = migrate.NewWithSourceInstance(
		"migrations", src,
		addr,
	); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Migrations) Up() error {
	if err := p.Migrate.Up(); !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
