INSERT INTO job (
    job_id,
    task_type,
    job_status
) VALUES (
    $1,
    $2,
    'WAITING'
) RETURNING job_id, task_type, job_status, job_error;
