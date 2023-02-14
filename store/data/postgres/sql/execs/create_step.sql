INSERT INTO step (
    step_id,
    job_id,
    task_type,
    job_args
) VALUES (
    $1,
    $2,
    $3,
    $4
) RETURNING step_id, job_id, task_type, job_args;
