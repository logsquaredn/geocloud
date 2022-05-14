INSERT INTO job (
    job_id,
    customer_id,
    input_id,
    output_id,
    task_type,
    job_args
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
) RETURNING job_id, customer_id, input_id, output_id, task_type, job_status, job_error, start_time, end_time, job_args;
