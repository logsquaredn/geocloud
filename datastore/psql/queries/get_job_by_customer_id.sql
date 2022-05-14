SELECT job_id, customer_id, input_id, output_id, task_type, job_status, job_error, start_time, end_time, job_args FROM job WHERE customer_id = $1;
