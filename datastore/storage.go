package datastore

import (
	"database/sql"
	_ "embed"
	"time"

	"github.com/google/uuid"
	"github.com/logsquaredn/geocloud"
)

var (
	//go:embed psql/execs/create_storage.sql
	createStorageSQL string

	//go:embed psql/execs/delete_storage.sql
	deleteStorageSQL string

	//go:embed psql/queries/get_storage_by_id.sql
	getStorageByIDSQL string

	//go:embed psql/execs/update_storage.sql
	updateStorageSQL string
)

func (p *Postgres) UpdateStorage(s *geocloud.Storage) (*geocloud.Storage, error) {
	var (
		lastUsed sql.NullTime
	)

	if err := p.stmt.createStorage.QueryRow(
		s.ID, time.Now(),
	).Scan(
		&s.ID, &s.CustomerID,
		&s.Name, &lastUsed,
	); err != nil {
		return s, err
	}

	s.LastUsed = lastUsed.Time

	return s, nil
}

func (p *Postgres) CreateStorage(s *geocloud.Storage) (*geocloud.Storage, error) {
	var (
		id       = uuid.NewString()
		lastUsed sql.NullTime
	)

	s.LastUsed = time.Now()
	if err := p.stmt.createStorage.QueryRow(
		id, s.CustomerID, s.Name,
	).Scan(
		&s.ID, &s.CustomerID,
		&s.Name, &lastUsed,
	); err != nil {
		return s, err
	}

	s.LastUsed = lastUsed.Time

	return s, nil
}

func (p *Postgres) GetStorage(m geocloud.Message) (*geocloud.Storage, error) {
	var (
		s        = &geocloud.Storage{}
		lastUsed sql.NullTime
	)

	if err := p.stmt.getStorage.QueryRow(m.GetID()).Scan(
		&s.ID, &s.CustomerID,
		&s.Name, &lastUsed,
	); err != nil {
		return s, err
	}

	s.LastUsed = lastUsed.Time

	return s, nil
}

func (p *Postgres) DeleteStorage(m geocloud.Message) error {
	_, err := p.stmt.deleteStorage.Exec(m.GetID())
	return err
}
