SELECT j.job_id, j.task_type, j.job_status, j.job_error, j.start_time, j.end_time, j.job_args, c.customer_id 
FROM job j
inner join customer c on j.job_id = c.job_id
where j.end_time < $1;
