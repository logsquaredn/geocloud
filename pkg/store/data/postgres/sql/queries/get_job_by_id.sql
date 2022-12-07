SELECT job_id, owner_id, input_id, output_id, job_status, job_error, start_time, end_time FROM job WHERE job_id = $1;
