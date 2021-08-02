SELECT DISTINCT task_queue_name FROM task where task.task_type = ANY($1);
