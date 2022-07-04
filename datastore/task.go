package datastore

import (
	"database/sql"
	_ "embed"

	"github.com/lib/pq"
	"github.com/logsquaredn/rototiller"
)

var (
	//go:embed psql/queries/get_task_by_job_id.sql
	getTaskByJobIDSQL string

	//go:embed psql/queries/get_tasks_by_types.sql
	getTasksByTypesSQL string
)

func (p *Postgres) GetTaskByJobID(m rototiller.Message) (*rototiller.Task, error) {
	var (
		t        = &rototiller.Task{}
		queueID  sql.NullString
		taskType string
		taskKind string
	)

	err := p.stmt.getTaskByJobID.QueryRow(m.GetID()).Scan(&taskType, &taskKind, pq.Array(&t.Params), &queueID)
	if err != nil {
		return t, err
	}

	t.QueueID = queueID.String
	t.Type, err = rototiller.ParseTaskType(taskType)
	if err != nil {
		return nil, err
	}

	t.Kind, err = rototiller.ParseTaskKind(taskKind)
	return t, err
}

//go:embed psql/queries/get_task_by_type.sql
var getTaskByTypeSQL string

func (p *Postgres) GetTask(tt rototiller.TaskType) (*rototiller.Task, error) {
	var (
		t        = &rototiller.Task{}
		queueID  sql.NullString
		taskType string
		taskKind string
	)
	err := p.stmt.getTaskByType.QueryRow(tt.String()).Scan(&taskType, &taskKind, pq.Array(&t.Params), &queueID)
	if err != nil {
		return t, err
	}

	t.QueueID = queueID.String
	t.Type, err = rototiller.ParseTaskType(taskType)
	if err != nil {
		return nil, err
	}

	t.Kind, err = rototiller.ParseTaskKind(taskKind)
	return t, err
}

func (p *Postgres) GetTasks(taskTypes ...rototiller.TaskType) ([]*rototiller.Task, error) {
	rawTaskTypes := make([]string, len(taskTypes))
	for i, tt := range taskTypes {
		rawTaskTypes[i] = tt.String()
	}

	rows, err := p.stmt.getTasksByTypes.Query(pq.Array(rawTaskTypes))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*rototiller.Task

	for rows.Next() {
		var (
			task     = &rototiller.Task{}
			queueID  sql.NullString
			taskType string
			taskKind string
		)

		if err = rows.Scan(&taskType, &taskKind, pq.Array(&task.Params), &queueID); err != nil {
			return nil, err
		}

		task.QueueID = queueID.String
		task.Type, err = rototiller.ParseTaskType(taskType)
		if err != nil {
			return nil, err
		}

		task.Kind, err = rototiller.ParseTaskKind(taskKind)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}
