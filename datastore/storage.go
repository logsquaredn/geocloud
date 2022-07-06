package datastore

import (
	_ "embed"
	"time"

	"github.com/google/uuid"
	"github.com/logsquaredn/rototiller"
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

	//go:embed psql/queries/get_storage_by_customer_id.sql
	getStorgageByCustomerIDSQL string

	//go:embed psql/queries/get_storage_before.sql
	getStorageBeforeSQL string

	//go:embed psql/queries/get_output_storage_by_job_id.sql
	getOutputStorageByJobIDSQL string

	//go:embed psql/queries/get_input_storage_by_job_id.sql
	getInputStorageByJobIDSQL string
)

func (p *Postgres) UpdateStorage(s *rototiller.Storage) (*rototiller.Storage, error) {
	var (
		storageStatus string
		err           error
	)
	if err := p.stmt.updateStorage.QueryRow(
		s.ID, s.Status, time.Now(),
	).Scan(
		&s.ID, &storageStatus, &s.CustomerID,
		&s.Name, &s.LastUsed, &s.CreateTime,
	); err != nil {
		return nil, err
	}

	s.Status, err = rototiller.ParseStorageStatus(storageStatus)
	return s, err
}

func (p *Postgres) CreateStorage(s *rototiller.Storage) (*rototiller.Storage, error) {
	var (
		id            = uuid.NewString()
		storageStatus string
		err           error
	)

	if s.Status == "" {
		s.Status = rototiller.StorageStatusUnknown
	}

	if err = p.stmt.createStorage.QueryRow(
		id, s.Status, s.CustomerID, s.Name,
	).Scan(
		&s.ID, &storageStatus, &s.CustomerID,
		&s.Name, &s.LastUsed, &s.CreateTime,
	); err != nil {
		return nil, err
	}

	s.Status, err = rototiller.ParseStorageStatus(storageStatus)
	return s, err
}

func (p *Postgres) GetStorage(m rototiller.Message) (*rototiller.Storage, error) {
	var (
		s             = &rototiller.Storage{}
		storageStatus string
		err           error
	)

	if err := p.stmt.getStorage.QueryRow(m.GetID()).Scan(
		&s.ID, &storageStatus, &s.CustomerID,
		&s.Name, &s.LastUsed, &s.CreateTime,
	); err != nil {
		return nil, err
	}

	s.Status, err = rototiller.ParseStorageStatus(storageStatus)
	return s, err
}

func (p *Postgres) DeleteStorage(m rototiller.Message) error {
	_, err := p.stmt.deleteStorage.Exec(m.GetID())
	return err
}

func (p *Postgres) GetCustomerStorage(m rototiller.Message, offset, limit int) ([]*rototiller.Storage, error) {
	rows, err := p.stmt.getStorageByCustomerID.Query(m.GetID(), offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	storage := []*rototiller.Storage{}

	for rows.Next() {
		var (
			s             = &rototiller.Storage{}
			storageStatus string
		)

		if err = rows.Scan(
			&s.ID, &storageStatus, &s.CustomerID,
			&s.Name, &s.LastUsed, &s.CreateTime,
		); err != nil {
			return nil, err
		}

		if s.Status, err = rototiller.ParseStorageStatus(storageStatus); err != nil {
			return nil, err
		}

		storage = append(storage, s)
	}

	return storage, nil
}

func (p *Postgres) GetJobInputStorage(m rototiller.Message) (*rototiller.Storage, error) {
	var (
		s             = &rototiller.Storage{}
		storageStatus string
		err           error
	)

	if err := p.stmt.getInputStorageByJobID.QueryRow(m.GetID()).Scan(
		&s.ID, &storageStatus, &s.CustomerID,
		&s.Name, &s.LastUsed, &s.CreateTime,
	); err != nil {
		return nil, err
	}

	s.Status, err = rototiller.ParseStorageStatus(storageStatus)
	return s, err
}

func (p *Postgres) GetJobOutputStorage(m rototiller.Message) (*rototiller.Storage, error) {
	var (
		s             = &rototiller.Storage{}
		storageStatus string
		err           error
	)

	if err := p.stmt.getOutputStorageByJobID.QueryRow(m.GetID()).Scan(
		&s.ID, &storageStatus, &s.CustomerID,
		&s.Name, &s.LastUsed, &s.CreateTime,
	); err != nil {
		return nil, err
	}

	s.Status, err = rototiller.ParseStorageStatus(storageStatus)
	return s, err
}

func (p *Postgres) GetStorageBefore(d time.Duration) ([]*rototiller.Storage, error) {
	beforeTimestamp := time.Now().Add(-d)
	rows, err := p.stmt.getStorageBefore.Query(beforeTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var storages []*rototiller.Storage

	for rows.Next() {
		var (
			s             = &rototiller.Storage{}
			storageStatus string
		)

		if err = rows.Scan(
			&s.ID, &storageStatus, &s.CustomerID,
			&s.Name, &s.LastUsed, &s.CreateTime,
		); err != nil {
			return nil, err
		}

		if s.Status, err = rototiller.ParseStorageStatus(storageStatus); err != nil {
			return nil, err
		}

		storages = append(storages, s)
	}

	return storages, nil
}
