INSERT INTO sink (
    sink_id,
    sink_address,
    job_id
) VALUES (
    $1,
    $2,
    $3
) RETURNING sink_id, sink_address, job_id;
