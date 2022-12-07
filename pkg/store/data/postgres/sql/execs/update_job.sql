UPDATE job SET (
    output_id,
    job_status,
    job_error,
    start_time,
    end_time
) = (
    $2,
    $3,
    $4,
    $5,
    $6
) WHERE job_id = $1 RETURNING job_id, owner_id, input_id, output_id, job_status, job_error, start_time, end_time;
