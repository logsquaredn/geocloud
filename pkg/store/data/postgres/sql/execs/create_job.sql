INSERT INTO job (
    job_id,
    owner_id,
    input_id
) VALUES (
    $1,
    $2,
    $3
) RETURNING job_id, owner_id, input_id, output_id, job_status, job_error, start_time, end_time;
