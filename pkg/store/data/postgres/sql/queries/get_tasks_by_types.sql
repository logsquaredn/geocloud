SELECT DISTINCT task_type, task_kind, task_params FROM task where task_type = ANY($1);
