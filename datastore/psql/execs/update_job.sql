UPDATE job SET (
    job_status,
    job_error,
    start_time,
    end_time
) = (
    $2,
    $3,
    $4,
    $5
) WHERE job_id = $1 RETURNING job_id, customer_id, input_id, output_id, task_type, job_status, job_error, start_time, end_time, job_args;
