package postgres

import (
	"database/sql"
	_ "embed"
	"time"

	"github.com/google/uuid"
	"github.com/logsquaredn/rototiller"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	//go:embed sql/execs/create_storage.sql
	createStorageSQL string

	//go:embed sql/execs/delete_storage.sql
	deleteStorageSQL string

	//go:embed sql/queries/get_storage_by_id.sql
	getStorageByIDSQL string

	//go:embed sql/execs/update_storage.sql
	updateStorageSQL string

	//go:embed sql/queries/get_storage_by_customer_id.sql
	getStorgageByCustomerIDSQL string

	//go:embed sql/queries/get_storage_before.sql
	getStorageBeforeSQL string

	//go:embed sql/queries/get_output_storage_by_job_id.sql
	getOutputStorageByJobIDSQL string

	//go:embed sql/queries/get_input_storage_by_job_id.sql
	getInputStorageByJobIDSQL string
)

func (d *Datastore) UpdateStorage(s *rototiller.Storage) (*rototiller.Storage, error) {
	var (
		lastUsed, createTime sql.NullTime
	)
	if err := d.stmt.updateStorage.QueryRow(
		s.Id, s.Status, time.Now(),
	).Scan(
		&s.Id, &s.Status, &s.CustomerId,
		&s.Name, &lastUsed, &createTime,
	); err != nil {
		return nil, err
	}

	s.LastUsed = timestamppb.New(lastUsed.Time)
	s.CreateTime = timestamppb.New(createTime.Time)

	return s, nil
}

func (d *Datastore) CreateStorage(s *rototiller.Storage) (*rototiller.Storage, error) {
	var (
		id                   = uuid.NewString()
		lastUsed, createTime sql.NullTime
		err                  error
	)

	if s.Status == "" {
		s.Status = rototiller.StorageStatusUnknown.String()
	}

	if err = d.stmt.createStorage.QueryRow(
		id, s.Status, s.CustomerId, s.Name,
	).Scan(
		&s.Id, &s.Status, &s.CustomerId,
		&s.Name, &lastUsed, &createTime,
	); err != nil {
		return nil, err
	}

	s.LastUsed = timestamppb.New(lastUsed.Time)
	s.CreateTime = timestamppb.New(createTime.Time)

	return s, nil
}

func (d *Datastore) GetStorage(id string) (*rototiller.Storage, error) {
	var (
		s                    = &rototiller.Storage{}
		lastUsed, createTime sql.NullTime
	)

	if err := d.stmt.getStorage.QueryRow(id).Scan(
		&s.Id, &s.Status, &s.CustomerId,
		&s.Name, &lastUsed, &createTime,
	); err != nil {
		return nil, err
	}

	s.LastUsed = timestamppb.New(lastUsed.Time)
	s.CreateTime = timestamppb.New(createTime.Time)

	return s, nil
}

func (d *Datastore) DeleteStorage(id string) error {
	_, err := d.stmt.deleteStorage.Exec(id)
	return err
}

func (d *Datastore) GetCustomerStorage(id string, offset, limit int) ([]*rototiller.Storage, error) {
	rows, err := d.stmt.getStorageByCustomerID.Query(id, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	storage := []*rototiller.Storage{}

	for rows.Next() {
		var (
			s                    = &rototiller.Storage{}
			lastUsed, createTime sql.NullTime
		)

		if err = rows.Scan(
			&s.Id, &s.Status, &s.CustomerId,
			&s.Name, &lastUsed, &createTime,
		); err != nil {
			return nil, err
		}

		s.LastUsed = timestamppb.New(lastUsed.Time)
		s.CreateTime = timestamppb.New(createTime.Time)

		storage = append(storage, s)
	}

	return storage, nil
}

func (d *Datastore) GetJobInputStorage(id string) (*rototiller.Storage, error) {
	var (
		s                    = &rototiller.Storage{}
		lastUsed, createTime sql.NullTime
	)

	if err := d.stmt.getInputStorageByJobID.QueryRow(id).Scan(
		&s.Id, &s.Status, &s.CustomerId,
		&s.Name, &lastUsed, &createTime,
	); err != nil {
		return nil, err
	}

	s.LastUsed = timestamppb.New(lastUsed.Time)
	s.CreateTime = timestamppb.New(createTime.Time)

	return s, nil
}

func (d *Datastore) GetJobOutputStorage(id string) (*rototiller.Storage, error) {
	var (
		s                    = &rototiller.Storage{}
		lastUsed, createTime sql.NullTime
	)

	if err := d.stmt.getOutputStorageByJobID.QueryRow(id).Scan(
		&s.Id, &s.Status, &s.CustomerId,
		&s.Name, &lastUsed, &createTime,
	); err != nil {
		return nil, err
	}

	s.LastUsed = timestamppb.New(lastUsed.Time)
	s.CreateTime = timestamppb.New(createTime.Time)

	return s, nil
}

func (d *Datastore) GetStorageBefore(duration time.Duration) ([]*rototiller.Storage, error) {
	beforeTimestamp := time.Now().Add(-duration)
	rows, err := d.stmt.getStorageBefore.Query(beforeTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var storages []*rototiller.Storage

	for rows.Next() {
		var (
			s                    = &rototiller.Storage{}
			lastUsed, createTime sql.NullTime
		)

		if err = rows.Scan(
			&s.Id, &s.Status, &s.CustomerId,
			&s.Name, &lastUsed, &createTime,
		); err != nil {
			return nil, err
		}

		s.LastUsed = timestamppb.New(lastUsed.Time)
		s.CreateTime = timestamppb.New(createTime.Time)

		storages = append(storages, s)
	}

	return storages, nil
}
