SELECT job_id, namespace, input_id, output_id, job_status, job_error, start_time, end_time
FROM job
WHERE end_time < $1;
