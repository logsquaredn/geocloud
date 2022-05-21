package datastore

import (
	"database/sql"
	_ "embed"

	"github.com/lib/pq"
	"github.com/logsquaredn/geocloud"
)

var (
	//go:embed psql/queries/get_task_by_job_id.sql
	getTaskByJobIDSQL string

	//go:embed psql/queries/get_tasks_by_types.sql
	getTasksByTypesSQL string
)

func (p *Postgres) GetTaskByJobID(m geocloud.Message) (*geocloud.Task, error) {
	var (
		t        = &geocloud.Task{}
		queueID  sql.NullString
		taskType string
	)

	err := p.stmt.getTaskByJobID.QueryRow(m.GetID()).Scan(&taskType, pq.Array(&t.Params), &queueID)
	if err != nil {
		return t, err
	}

	t.QueueID = queueID.String
	t.Type, err = geocloud.TaskTypeFrom(taskType)
	return t, err
}

//go:embed psql/queries/get_task_by_type.sql
var getTaskByTypeSQL string

func (p *Postgres) GetTask(tt geocloud.TaskType) (*geocloud.Task, error) {
	var (
		t        = &geocloud.Task{}
		queueID  sql.NullString
		taskType string
	)
	err := p.stmt.getTaskByType.QueryRow(tt.String()).Scan(&taskType, pq.Array(&t.Params), &queueID)
	if err != nil {
		return t, err
	}

	t.QueueID = queueID.String
	t.Type, err = geocloud.TaskTypeFrom(taskType)
	return t, err
}

func (p *Postgres) GetTasks(taskTypes ...geocloud.TaskType) ([]*geocloud.Task, error) {
	rawTaskTypes := make([]string, len(taskTypes))
	for i, tt := range taskTypes {
		rawTaskTypes[i] = tt.String()
	}

	rows, err := p.stmt.getTasksByTypes.Query(pq.Array(rawTaskTypes))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*geocloud.Task

	for rows.Next() {
		var (
			task     = &geocloud.Task{}
			queueID  sql.NullString
			taskType string
		)

		err = rows.Scan(&taskType, pq.Array(&task.Params), &queueID)
		if err != nil {
			return nil, err
		}

		task.QueueID = queueID.String
		task.Type, err = geocloud.TaskTypeFrom(taskType)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}
