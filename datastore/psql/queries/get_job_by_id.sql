SELECT job_id, task_type, job_status, job_error, start_time, end_time, job_args FROM job WHERE job_id = $1;
