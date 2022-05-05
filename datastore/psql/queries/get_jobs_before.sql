SELECT j.job_id, j.task_type, j.job_status, j.job_error, j.start_time, j.end_time, j.job_args, jcm.customer_id 
FROM job j
inner join job_customer_mapping jcm on j.job_id = jcm.job_id
where j.end_time < $1 and j.job_status = 'complete' or j.job_status = 'error';
