UPDATE job SET (
    job_status,
    job_error,
    end_time
) = (
    'ERROR',
    $2,
    NOW()
) where job_id = $1 RETURNING job_id, task_type, job_status, job_error, start_time, end_time, job_params;
