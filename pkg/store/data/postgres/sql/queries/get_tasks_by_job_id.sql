SELECT t.task_type, task_kind, task_params 
FROM task t INNER JOIN step s ON t.task_type = s.task_type 
WHERE s.job_id = $1;
