SELECT task_type, task_params, task_queue_name, task_ref FROM task where task_type = $1;
