UPDATE job SET (
    task_type,
    job_status,
    job_error,
    start_time,
    end_time,
    job_args
) = (
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
) where job_id = $1 RETURNING job_id, task_type, job_status, job_error, start_time, end_time, job_args;
