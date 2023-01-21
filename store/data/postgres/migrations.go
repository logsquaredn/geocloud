package postgres

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

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

	if p.Migrate, err = migrate.NewWithSourceInstance(
		"migrations", src,
		u.String(),
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
