package postgres

import (
	"context"
	_ "embed"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/logsquaredn/rototiller/pkg/api"
	"golang.org/x/sync/errgroup"
)

var (
	//go:embed sql/execs/create_step.sql
	createStepSQL string

	//go:embed sql/queries/get_steps_by_job_id.sql
	getStepsByJobIDSQL string
)

func (d *Datastore) createSteps(jobID string, steps []*api.Step) ([]*api.Step, error) {
	eg, _ := errgroup.WithContext(context.TODO())
	for i, step := range steps {
		i := i
		step := step
		eg.Go(func() error {
			id := uuid.New().String()
			return d.stmt.createStep.QueryRow(
				id, jobID,
				step.TaskType,
				step.Args,
			).Scan(
				&step.Id, &step.JobId,
				&step.TaskType, pq.Array(&step.Args),
			)
		})

		steps[i] = step
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return steps, nil
}

func (d *Datastore) getSteps(jobID string) ([]*api.Step, error) {
	rows, err := d.stmt.getStepsByJobID.Query(jobID)
	if err != nil {
		return nil, err
	}

	var steps []*api.Step
	for rows.Next() {
		s := &api.Step{}
		err = rows.Scan(
			&s.Id, &s.JobId,
			&s.TaskType, pq.Array(&s.Args),
		)
		if err != nil {
			return nil, err
		}

		steps = append(steps, s)
	}

	return steps, nil
}
