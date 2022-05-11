SELECT j.job_id, j.task_type, j.job_status, j.job_error, j.start_time, j.end_time, j.job_args, jcm.customer_id, c.customer_name
FROM job j
inner join job_customer_mapping jcm on j.job_id = jcm.job_id
inner join customer c on jcm.customer_id = c.customer_id
where j.end_time < $1;
