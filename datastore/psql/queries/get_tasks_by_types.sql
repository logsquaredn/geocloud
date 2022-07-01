SELECT DISTINCT task_type, task_kind, task_params, task_queue_id FROM task where task_type = ANY($1);
