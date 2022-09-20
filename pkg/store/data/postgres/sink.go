package postgres

import (
	_ "embed"

	"github.com/google/uuid"
	"github.com/logsquaredn/rototiller/pkg/api"
)

var (
	//go:embed sql/execs/create_sink.sql
	createSinkSQL string

	//go:embed sql/queries/get_sinks_by_job_id.sql
	getSinksByJobIDSQL string
)

func (d *Datastore) CreateSink(s *api.Sink) (*api.Sink, error) {
	return s, d.stmt.createSink.QueryRow(
		uuid.NewString(), s.Address,
		s.JobId,
	).Scan(
		&s.Id, &s.Address,
		&s.JobId,
	)
}

func (d *Datastore) GetJobSinks(id string) ([]*api.Sink, error) {
	rows, err := d.stmt.getSinksByJobID.Query(id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sinks := []*api.Sink{}
	for rows.Next() {
		s := &api.Sink{}

		if err = rows.Scan(
			&s.Id, &s.Address,
			&s.JobId,
		); err != nil {
			return nil, err
		}

		sinks = append(sinks, s)
	}

	return sinks, nil
}
