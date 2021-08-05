SELECT DISTINCT task_ref FROM task where task.task_type = ANY($1);
