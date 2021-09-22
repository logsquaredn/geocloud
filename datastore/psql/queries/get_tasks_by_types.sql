SELECT DISTINCT task_type, task_params, task_queue_id, task_ref FROM task where task_type = ANY($1);
