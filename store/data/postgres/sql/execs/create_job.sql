INSERT INTO job (
    job_id,
    namespace,
    input_id,
    task_type,
    job_args
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
) RETURNING job_id, namespace, input_id, output_id, task_type, job_status, job_error, start_time, end_time, job_args;
