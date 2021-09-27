SELECT t.task_type, task_params, task_queue_id, task_ref FROM task t INNER JOIN job j ON t.task_type = j.task_type WHERE j.job_id = $1;
