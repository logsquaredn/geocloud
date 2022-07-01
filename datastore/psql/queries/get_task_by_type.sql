SELECT task_type, task_kind, task_params, task_queue_id FROM task where task_type = $1;
