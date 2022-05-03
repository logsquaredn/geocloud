SELECT job_id, task_type, job_status, job_error, start_time, end_time, job_args FROM job where end_time < $1;
