package postgres

import (
	_ "embed"

	"github.com/lib/pq"
	"github.com/logsquaredn/rototiller"
)

var (
	//go:embed sql/queries/get_tasks_by_job_id.sql
	getTasksByJobIDSQL string

	//go:embed sql/queries/get_tasks_by_types.sql
	getTasksByTypesSQL string
)

func (d *Datastore) GetTasksByJobID(id string) ([]*rototiller.Task, error) {
	rows, err := d.stmt.getTasksByJobID.Query(id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*rototiller.Task

	for rows.Next() {
		t := &rototiller.Task{}

		if err = rows.Scan(&t.Type, &t.Kind, pq.Array(&t.Params)); err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

//go:embed sql/queries/get_task_by_type.sql
var getTaskByTypeSQL string

func (d *Datastore) GetTask(tt rototiller.TaskType) (*rototiller.Task, error) {
	t := &rototiller.Task{}

	if err := d.stmt.getTaskByType.QueryRow(tt.String()).Scan(&t.Type, &t.Kind, pq.Array(&t.Params)); err != nil {
		return nil, err
	}

	return t, nil
}

func (d *Datastore) GetTasks(taskTypes ...rototiller.TaskType) ([]*rototiller.Task, error) {
	rawTaskTypes := make([]string, len(taskTypes))
	for i, tt := range taskTypes {
		rawTaskTypes[i] = tt.String()
	}

	rows, err := d.stmt.getTasksByTypes.Query(pq.Array(rawTaskTypes))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*rototiller.Task

	for rows.Next() {
		t := &rototiller.Task{}

		if err = rows.Scan(&t.Type, &t.Kind, pq.Array(&t.Params)); err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}
